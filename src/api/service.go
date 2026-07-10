package api

import (
	"github.com/solguardlabs/northstardtl/src/audit"
	"github.com/solguardlabs/northstardtl/src/domain"
	"github.com/solguardlabs/northstardtl/src/ledger"
	"github.com/solguardlabs/northstardtl/src/routing"
	"github.com/solguardlabs/northstardtl/src/settlement"
)

type Service struct {
	epoch    int64
	book     *ledger.Book
	catalog  *routing.Catalog
	scorer   routing.Scorer
	selector routing.Selector
	events   *domain.EventSink
	engine   *settlement.Engine
}

func NewService(bootstrap Bootstrap) (*Service, error) {
	bootstrap = bootstrap.WithDefaults()
	if len(bootstrap.Routes) == 0 {
		return nil, domain.NewError(domain.ErrInvalidBootstrap, "at least one route is required")
	}
	seeds := append(ledger.RouteSeeds(bootstrap.Routes), bootstrap.Balances...)
	book, err := ledger.NewBook(seeds)
	if err != nil {
		return nil, err
	}
	catalog, err := routing.NewCatalog(bootstrap.Routes)
	if err != nil {
		return nil, err
	}
	scorer := routing.NewScorer(bootstrap.Policy)
	selector := routing.NewSelector(scorer)
	events := domain.NewEventSink()
	service := &Service{
		epoch:    bootstrap.Epoch,
		book:     book,
		catalog:  catalog,
		scorer:   scorer,
		selector: selector,
		events:   events,
	}
	service.engine = settlement.NewEngine(book, catalog, selector, events)
	return service, nil
}

func (service *Service) Epoch() int64 {
	return service.epoch
}

func (service *Service) QuoteIntent(intent domain.Intent) (domain.QuoteBook, error) {
	if err := intent.Validate(); err != nil {
		return domain.QuoteBook{}, err
	}
	book := service.selector.Quotes(intent, service.catalog.Matching(intent), service.epoch)
	if len(book.Quotes) > 0 {
		service.events.Append(domain.Event{
			Type:    domain.EventRouteQuoted,
			Epoch:   service.epoch,
			RouteID: book.Quotes[0].RouteID,
			Intent:  intent.ID,
			Message: "intent quoted against route catalog",
		})
	}
	return book, nil
}

func (service *Service) SubmitIntent(request domain.SubmitIntentRequest) (domain.SubmitIntentResponse, error) {
	return service.engine.Submit(request.Intent, service.epoch)
}

func (service *Service) Execute(request domain.ExecuteRequest) (domain.ExecuteResponse, error) {
	return service.engine.Execute(request, service.epoch)
}

func (service *Service) AdvanceEpoch(delta int64) int64 {
	if delta <= 0 {
		delta = 1
	}
	service.epoch += delta
	service.events.Append(domain.Event{
		Type:    domain.EventEpochAdvanced,
		Epoch:   service.epoch,
		Message: "settlement epoch advanced",
	})
	return service.epoch
}

func (service *Service) UpdateRoute(patch domain.RoutePatch) (domain.Route, error) {
	route, err := service.catalog.ApplyPatch(patch)
	if err != nil {
		return domain.Route{}, err
	}
	service.events.Append(domain.Event{
		Type:    domain.EventRouteUpdated,
		Epoch:   service.epoch,
		RouteID: route.ID,
		Message: "route configuration updated",
	})
	return route, nil
}

func (service *Service) ApplyCongestion(delta routing.CongestionDelta) (domain.Route, error) {
	route, err := routing.ApplyCongestion(service.catalog, delta)
	if err != nil {
		return domain.Route{}, err
	}
	service.events.Append(domain.Event{
		Type:    domain.EventRouteUpdated,
		Epoch:   service.epoch,
		RouteID: route.ID,
		Message: "route congestion updated",
	})
	return route, nil
}

func (service *Service) Snapshot() domain.Snapshot {
	snapshot := domain.Snapshot{
		Epoch:    service.epoch,
		Balances: service.book.Snapshot(),
		Routes:   service.catalog.Snapshot(),
		Queue:    service.engine.QueueSnapshot(),
		Receipts: service.engine.Receipts(),
		Events:   service.events.Snapshot(),
	}
	snapshot.AuditIssues = audit.Check(service.book, service.catalog.Routes(), service.engine.Tickets())
	return snapshot
}

func (service *Service) Policy() routing.Policy {
	return service.scorer.Policy()
}
