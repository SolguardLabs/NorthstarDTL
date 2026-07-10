package routing

import (
	"sort"

	"github.com/solguardlabs/northstardtl/src/domain"
)

type Catalog struct {
	routes map[domain.RouteID]domain.Route
}

func NewCatalog(routes []domain.Route) (*Catalog, error) {
	catalog := &Catalog{routes: make(map[domain.RouteID]domain.Route, len(routes))}
	for _, route := range routes {
		if err := catalog.Add(route); err != nil {
			return nil, err
		}
	}
	return catalog, nil
}

func (catalog *Catalog) Add(route domain.Route) error {
	if err := route.Validate(); err != nil {
		return err
	}
	if _, ok := catalog.routes[route.ID]; ok {
		return domain.Duplicate("route %s already exists", route.ID)
	}
	catalog.routes[route.ID] = route
	return nil
}

func (catalog *Catalog) Get(id domain.RouteID) (domain.Route, bool) {
	route, ok := catalog.routes[id]
	return route, ok
}

func (catalog *Catalog) MustGet(id domain.RouteID) (domain.Route, error) {
	route, ok := catalog.Get(id)
	if !ok {
		return domain.Route{}, domain.NewError(domain.ErrRouteNotFound, "route %s not found", id)
	}
	return route, nil
}

func (catalog *Catalog) ApplyPatch(patch domain.RoutePatch) (domain.Route, error) {
	route, ok := catalog.routes[patch.ID]
	if !ok {
		return domain.Route{}, domain.NewError(domain.ErrRouteNotFound, "route %s not found", patch.ID)
	}
	route = route.ApplyPatch(patch)
	if err := route.Validate(); err != nil {
		return domain.Route{}, err
	}
	catalog.routes[route.ID] = route
	return route, nil
}

func (catalog *Catalog) UpdateExposure(id domain.RouteID, exposureDelta domain.Money, liquidityDelta domain.Money) (domain.Route, error) {
	route, ok := catalog.routes[id]
	if !ok {
		return domain.Route{}, domain.NewError(domain.ErrRouteNotFound, "route %s not found", id)
	}
	if exposureDelta < 0 && route.Exposure < -exposureDelta {
		return domain.Route{}, domain.Invalid("route %s exposure cannot go below zero", id)
	}
	if liquidityDelta < 0 && route.OutputLiquidity < -liquidityDelta {
		return domain.Route{}, domain.InsufficientRoute("route %s output liquidity would go below zero", id)
	}
	route.Exposure += exposureDelta
	route.OutputLiquidity += liquidityDelta
	if err := route.Validate(); err != nil {
		return domain.Route{}, err
	}
	catalog.routes[id] = route
	return route, nil
}

func (catalog *Catalog) Routes() []domain.Route {
	routes := make([]domain.Route, 0, len(catalog.routes))
	for _, route := range catalog.routes {
		routes = append(routes, route)
	}
	sort.Slice(routes, func(i, j int) bool {
		return routes[i].ID < routes[j].ID
	})
	return routes
}

func (catalog *Catalog) Matching(intent domain.Intent) []domain.Route {
	routes := make([]domain.Route, 0, len(catalog.routes))
	for _, route := range catalog.routes {
		if route.PairMatches(intent) {
			routes = append(routes, route)
		}
	}
	sort.Slice(routes, func(i, j int) bool {
		if routes[i].Preferred != routes[j].Preferred {
			return routes[i].Preferred
		}
		return routes[i].ID < routes[j].ID
	})
	return routes
}

func (catalog *Catalog) Snapshot() []domain.RouteSnapshot {
	routes := catalog.Routes()
	out := make([]domain.RouteSnapshot, len(routes))
	for index, route := range routes {
		out[index] = domain.NewRouteSnapshot(route)
	}
	return out
}

func (catalog *Catalog) TotalExposure(asset domain.Asset) domain.Money {
	var total domain.Money
	for _, route := range catalog.routes {
		if route.SourceAsset == asset {
			total += route.Exposure
		}
	}
	return total
}
