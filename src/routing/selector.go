package routing

import (
	"sort"

	"github.com/solguardlabs/northstardtl/src/domain"
)

type Selector struct {
	scorer Scorer
}

func NewSelector(scorer Scorer) Selector {
	return Selector{scorer: scorer}
}

func (selector Selector) Quotes(intent domain.Intent, routes []domain.Route, now int64) domain.QuoteBook {
	quotes := make([]domain.Quote, 0, len(routes))
	for _, route := range routes {
		quotes = append(quotes, selector.scorer.Quote(route, intent, now))
	}
	sort.Slice(quotes, func(i, j int) bool {
		return quotes[i].BetterThan(quotes[j])
	})
	return domain.QuoteBook{IntentID: intent.ID, Quotes: quotes}
}

func (selector Selector) Select(intent domain.Intent, routes []domain.Route, now int64) (domain.Quote, error) {
	book := selector.Quotes(intent, routes, now)
	for _, quote := range book.Quotes {
		route, ok := routeByID(routes, quote.RouteID)
		if ok && selector.scorer.Admissible(route, intent, quote) {
			return quote, nil
		}
	}
	if len(book.Quotes) == 0 {
		return domain.Quote{}, domain.NewError(domain.ErrRouteUnavailable, "no route supports %s/%s", intent.SourceAsset, intent.DestinationAsset)
	}
	return domain.Quote{}, domain.NewError(domain.ErrRouteUnavailable, "no route currently accepts intent %s", intent.ID)
}

func (selector Selector) SelectFallback(ticket domain.Ticket, routes []domain.Route, now int64) (domain.Route, error) {
	candidates := make([]fallbackCandidate, 0, len(routes))
	for _, route := range routes {
		if route.ID == ticket.ActiveRouteID {
			continue
		}
		if !route.PairMatches(ticket.Intent) {
			continue
		}
		if !selector.scorer.FallbackAdmissible(route, ticket.Intent, ticket.Quote.NetOut) {
			continue
		}
		quote := selector.scorer.Quote(route, ticket.Intent, now)
		candidates = append(candidates, fallbackCandidate{route: route, quote: quote})
	}
	sort.Slice(candidates, func(i, j int) bool {
		if candidates[i].quote.Score != candidates[j].quote.Score {
			return candidates[i].quote.Score > candidates[j].quote.Score
		}
		if candidates[i].route.Preferred != candidates[j].route.Preferred {
			return candidates[i].route.Preferred
		}
		return candidates[i].route.ID < candidates[j].route.ID
	})
	if len(candidates) == 0 {
		return domain.Route{}, domain.NewError(domain.ErrRouteUnavailable, "no fallback route accepts ticket %s", ticket.ID)
	}
	return candidates[0].route, nil
}

func (selector Selector) ShouldFallback(ticket domain.Ticket, route domain.Route) bool {
	if !route.ActiveForAdmission() {
		return true
	}
	if route.HealthBps < selector.scorer.policy.MinHealthBps {
		return true
	}
	if route.CongestionBps > selector.scorer.policy.MaxFallbackCongestion {
		return true
	}
	if route.RemainingExposure() < ticket.Intent.Amount {
		return true
	}
	if route.OutputLiquidity < ticket.Quote.NetOut {
		return true
	}
	return false
}

type fallbackCandidate struct {
	route domain.Route
	quote domain.Quote
}

func routeByID(routes []domain.Route, id domain.RouteID) (domain.Route, bool) {
	for _, route := range routes {
		if route.ID == id {
			return route, true
		}
	}
	return domain.Route{}, false
}
