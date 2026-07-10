package settlement

import (
	"sort"

	"github.com/solguardlabs/northstardtl/src/domain"
)

type JournalEntryType string

const (
	JournalAdmission JournalEntryType = "admission"
	JournalExecution JournalEntryType = "execution"
	JournalFallback  JournalEntryType = "fallback"
	JournalDeferred  JournalEntryType = "deferred"
)

type JournalEntry struct {
	Type         JournalEntryType `json:"type"`
	TicketID     domain.TicketID  `json:"ticketId"`
	IntentID     domain.IntentID  `json:"intentId"`
	RouteID      domain.RouteID   `json:"routeId"`
	QuotedRoute  domain.RouteID   `json:"quotedRouteId,omitempty"`
	Epoch        int64            `json:"epoch"`
	Amount       domain.Money     `json:"amount"`
	Destination  domain.Money     `json:"destination"`
	Fee          domain.Money     `json:"fee"`
	QueueScore   int64            `json:"queueScore"`
	UsedFallback bool             `json:"usedFallback"`
}

type Journal struct {
	Entries []JournalEntry `json:"entries"`
}

func BuildJournal(tickets []domain.Ticket, receipts []domain.SettlementReceipt) Journal {
	entries := make([]JournalEntry, 0, len(tickets)+len(receipts))
	for _, ticket := range tickets {
		entryType := JournalAdmission
		if ticket.Status == domain.TicketReady || ticket.Status == domain.TicketQueued {
			entryType = JournalDeferred
		}
		entries = append(entries, JournalEntry{
			Type:        entryType,
			TicketID:    ticket.ID,
			IntentID:    ticket.Intent.ID,
			RouteID:     ticket.ActiveRouteID,
			QuotedRoute: ticket.Quote.RouteID,
			Epoch:       ticket.LastTransition,
			Amount:      ticket.Intent.Amount,
			Destination: ticket.Quote.NetOut,
			Fee:         ticket.Quote.RouteFee,
			QueueScore:  ticket.QueueScore,
		})
	}
	for _, receipt := range receipts {
		entryType := JournalExecution
		if receipt.UsedFallback {
			entryType = JournalFallback
		}
		entries = append(entries, JournalEntry{
			Type:         entryType,
			TicketID:     receipt.TicketID,
			IntentID:     receipt.IntentID,
			RouteID:      receipt.RouteID,
			QuotedRoute:  receipt.QuotedRouteID,
			Epoch:        receipt.ExecutedEpoch,
			Amount:       receipt.AmountIn,
			Destination:  receipt.DestinationAmount,
			Fee:          receipt.RouteFee,
			QueueScore:   receipt.QueueAttempts,
			UsedFallback: receipt.UsedFallback,
		})
	}
	sort.SliceStable(entries, func(i, j int) bool {
		if entries[i].Epoch != entries[j].Epoch {
			return entries[i].Epoch < entries[j].Epoch
		}
		if entries[i].TicketID != entries[j].TicketID {
			return entries[i].TicketID < entries[j].TicketID
		}
		return entries[i].Type < entries[j].Type
	})
	return Journal{Entries: entries}
}

func (journal Journal) ByTicket(ticketID domain.TicketID) []JournalEntry {
	out := make([]JournalEntry, 0)
	for _, entry := range journal.Entries {
		if entry.TicketID == ticketID {
			out = append(out, entry)
		}
	}
	return out
}

func (journal Journal) Totals() domain.SettlementStats {
	stats := domain.SettlementStats{}
	for _, entry := range journal.Entries {
		if entry.Type != JournalExecution && entry.Type != JournalFallback {
			continue
		}
		stats.ReceiptCount++
		if entry.UsedFallback {
			stats.FallbackCount++
		}
		stats.TotalSourceAmount += entry.Amount
		stats.TotalDestAmount += entry.Destination
		stats.TotalFees += entry.Fee
	}
	return stats
}

func (journal Journal) Fallbacks() []JournalEntry {
	out := make([]JournalEntry, 0)
	for _, entry := range journal.Entries {
		if entry.Type == JournalFallback {
			out = append(out, entry)
		}
	}
	return out
}
