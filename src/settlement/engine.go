package settlement

import (
	"fmt"

	"github.com/solguardlabs/northstardtl/src/domain"
	"github.com/solguardlabs/northstardtl/src/ledger"
	"github.com/solguardlabs/northstardtl/src/routing"
)

type Engine struct {
	book       *ledger.Book
	catalog    *routing.Catalog
	selector   routing.Selector
	events     *domain.EventSink
	queue      *Queue
	receipts   []domain.SettlementReceipt
	ticketSeq  int64
	receiptSeq int64
}

func NewEngine(book *ledger.Book, catalog *routing.Catalog, selector routing.Selector, events *domain.EventSink) *Engine {
	return &Engine{
		book:     book,
		catalog:  catalog,
		selector: selector,
		events:   events,
		queue:    NewQueue(),
	}
}

func (engine *Engine) Submit(intent domain.Intent, now int64) (domain.SubmitIntentResponse, error) {
	if err := intent.Validate(); err != nil {
		return domain.SubmitIntentResponse{}, err
	}
	quote, err := engine.selector.Select(intent, engine.catalog.Matching(intent), now)
	if err != nil {
		return domain.SubmitIntentResponse{}, err
	}
	if err := engine.book.Reserve(intent.SourceAccount, intent.SourceAsset, quote.ReserveAmount(), "route admission reserve"); err != nil {
		return domain.SubmitIntentResponse{}, err
	}
	engine.ticketSeq++
	ticket := domain.NewTicket(domain.NewTicketID(intent.ID, engine.ticketSeq), intent, quote, now)
	if err := engine.queue.Add(ticket); err != nil {
		return domain.SubmitIntentResponse{}, err
	}
	engine.events.Append(domain.Event{
		Type:    domain.EventTicketSubmitted,
		Epoch:   now,
		RouteID: quote.RouteID,
		Ticket:  ticket.ID,
		Intent:  intent.ID,
		Message: "intent admitted into async settlement queue",
	})
	return domain.SubmitIntentResponse{Ticket: ticket, Quote: quote}, nil
}

func (engine *Engine) Execute(request domain.ExecuteRequest, now int64) (domain.ExecuteResponse, error) {
	ready := engine.queue.Ready(now, request.Count)
	response := domain.ExecuteResponse{
		Receipts: make([]domain.SettlementReceipt, 0, len(ready)),
		Deferred: make([]domain.Ticket, 0),
	}
	for _, ticket := range ready {
		executed, receipt, err := engine.executeOne(ticket, now)
		if err != nil {
			if converted, ok := domain.AsDomainError(err); ok && converted.Code == domain.ErrRouteUnavailable {
				ticket.Attempts++
				_ = engine.queue.Update(ticket)
				response.Deferred = append(response.Deferred, ticket)
				engine.events.Append(domain.Event{
					Type:    domain.EventTicketDeferred,
					Epoch:   now,
					Ticket:  ticket.ID,
					Intent:  ticket.Intent.ID,
					Message: converted.Message,
				})
				continue
			}
			return domain.ExecuteResponse{}, err
		}
		if executed {
			response.Receipts = append(response.Receipts, receipt)
		}
	}
	return response, nil
}

func (engine *Engine) executeOne(ticket domain.Ticket, now int64) (bool, domain.SettlementReceipt, error) {
	ticket.Attempts++
	route, ok := engine.catalog.Get(ticket.ActiveRouteID)
	if !ok {
		return false, domain.SettlementReceipt{}, domain.NewError(domain.ErrRouteUnavailable, "active route %s not found", ticket.ActiveRouteID)
	}
	if engine.selector.ShouldFallback(ticket, route) {
		fallback, err := engine.selector.SelectFallback(ticket, engine.catalog.Matching(ticket.Intent), now)
		if err != nil {
			return false, domain.SettlementReceipt{}, err
		}
		previous := ticket.ActiveRouteID
		ticket = ticket.WithActiveRoute(fallback.ID, now)
		route = fallback
		engine.events.Append(domain.Event{
			Type:    domain.EventFallbackApplied,
			Epoch:   now,
			RouteID: route.ID,
			Ticket:  ticket.ID,
			Intent:  ticket.Intent.ID,
			Message: fmt.Sprintf("ticket moved from %s to %s", previous, route.ID),
		})
	}
	if err := engine.preflight(ticket, route); err != nil {
		return false, domain.SettlementReceipt{}, err
	}
	receipt, err := engine.settle(ticket, route, now)
	if err != nil {
		return false, domain.SettlementReceipt{}, err
	}
	ticket = ticket.MarkExecuted(now)
	if err := engine.queue.Update(ticket); err != nil {
		return false, domain.SettlementReceipt{}, err
	}
	engine.receipts = append(engine.receipts, receipt)
	engine.events.Append(domain.Event{
		Type:    domain.EventTicketExecuted,
		Epoch:   now,
		RouteID: receipt.RouteID,
		Ticket:  ticket.ID,
		Intent:  ticket.Intent.ID,
		Message: "ticket settled",
	})
	return true, receipt, nil
}

func (engine *Engine) preflight(ticket domain.Ticket, route domain.Route) error {
	if ticket.Status == domain.TicketExecuted || ticket.Status == domain.TicketRejected {
		return domain.TicketState("ticket %s is not pending", ticket.ID)
	}
	if !route.PairMatches(ticket.Intent) {
		return domain.NewError(domain.ErrRouteUnavailable, "route %s no longer supports ticket %s pair", route.ID, ticket.ID)
	}
	if !route.ActiveForFallback() {
		return domain.NewError(domain.ErrRouteUnavailable, "route %s is not executable", route.ID)
	}
	if route.RemainingExposure() < ticket.Intent.Amount {
		return domain.ExposureLimit("route %s exposure limit exceeded", route.ID)
	}
	if route.OutputLiquidity < ticket.Quote.NetOut {
		return domain.InsufficientRoute("route %s output liquidity below ticket destination", route.ID)
	}
	available := engine.book.Balance(route.DestinationVault, ticket.Intent.DestinationAsset).Available
	if available < ticket.Quote.NetOut {
		return domain.InsufficientRoute("route %s vault has %s available, needs %s", route.ID, available, ticket.Quote.NetOut)
	}
	reserved := engine.book.Balance(ticket.Intent.SourceAccount, ticket.Intent.SourceAsset).Reserved
	if reserved < ticket.Quote.ReserveAmount() {
		return domain.InsufficientFunds("ticket %s reserve is below quoted amount", ticket.ID)
	}
	return nil
}

func (engine *Engine) settle(ticket domain.Ticket, route domain.Route, now int64) (domain.SettlementReceipt, error) {
	intent := ticket.Intent
	quote := ticket.Quote
	memo := fmt.Sprintf("settlement %s", ticket.ID)
	if err := engine.book.SpendReserved(intent.SourceAccount, intent.SourceAsset, quote.ReserveAmount(), memo); err != nil {
		return domain.SettlementReceipt{}, err
	}
	if err := engine.book.Credit(route.SourceTreasury, intent.SourceAsset, quote.AmountIn, memo); err != nil {
		return domain.SettlementReceipt{}, err
	}
	if err := engine.book.Credit(route.FeeAccount, intent.SourceAsset, quote.RouteFee, memo); err != nil {
		return domain.SettlementReceipt{}, err
	}
	if err := engine.book.Debit(route.DestinationVault, intent.DestinationAsset, quote.NetOut, memo); err != nil {
		return domain.SettlementReceipt{}, err
	}
	if err := engine.book.Credit(intent.DestinationAccount, intent.DestinationAsset, quote.NetOut, memo); err != nil {
		return domain.SettlementReceipt{}, err
	}
	updated, err := engine.catalog.UpdateExposure(route.ID, quote.AmountIn, -quote.NetOut)
	if err != nil {
		return domain.SettlementReceipt{}, err
	}
	engine.receiptSeq++
	receipt := domain.SettlementReceipt{
		ID:                   domain.NewReceiptID(ticket.ID, engine.receiptSeq),
		TicketID:             ticket.ID,
		IntentID:             intent.ID,
		RouteID:              route.ID,
		QuotedRouteID:        quote.RouteID,
		SourceAccount:        intent.SourceAccount,
		DestinationAccount:   intent.DestinationAccount,
		SourceAsset:          intent.SourceAsset,
		DestinationAsset:     intent.DestinationAsset,
		AmountIn:             quote.AmountIn,
		DestinationAmount:    quote.NetOut,
		RouteFee:             quote.RouteFee,
		CongestionPenalty:    quote.CongestionPenalty,
		ExecutedEpoch:        now,
		UsedFallback:         route.ID != quote.RouteID,
		FallbackFrom:         ticket.FallbackFrom,
		RouteExposureAfter:   updated.Exposure,
		RouteLiquidityAfter:  updated.OutputLiquidity,
		QueueAttempts:        ticket.Attempts,
		SettlementLedgerMemo: memo,
	}
	return receipt, nil
}

func (engine *Engine) QueueSnapshot() []domain.QueueSnapshot {
	return engine.queue.Snapshot()
}

func (engine *Engine) Receipts() []domain.SettlementReceipt {
	out := make([]domain.SettlementReceipt, len(engine.receipts))
	copy(out, engine.receipts)
	return out
}

func (engine *Engine) Tickets() []domain.Ticket {
	return engine.queue.All()
}
