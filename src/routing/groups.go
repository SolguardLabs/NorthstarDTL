package routing

import (
	"sort"

	"github.com/solguardlabs/northstardtl/src/domain"
)

type RouteGroup struct {
	Pair          domain.PairKey   `json:"pair"`
	RouteIDs      []domain.RouteID `json:"routeIds"`
	Active        int              `json:"active"`
	TotalExposure domain.Money     `json:"totalExposure"`
	TotalOutput   domain.Money     `json:"totalOutput"`
	BestRatePpm   int64            `json:"bestRatePpm"`
	WorstRatePpm  int64            `json:"worstRatePpm"`
}

func GroupRoutes(routes []domain.Route) []RouteGroup {
	index := make(map[domain.PairKey]*RouteGroup)
	for _, route := range routes {
		pair := domain.PairOf(route)
		group := index[pair]
		if group == nil {
			group = &RouteGroup{
				Pair:         pair,
				RouteIDs:     make([]domain.RouteID, 0),
				BestRatePpm:  route.RatePpm,
				WorstRatePpm: route.RatePpm,
			}
			index[pair] = group
		}
		group.RouteIDs = append(group.RouteIDs, route.ID)
		if route.StatusOrDefault() == domain.RouteActive {
			group.Active++
		}
		group.TotalExposure += route.Exposure
		group.TotalOutput += route.OutputLiquidity
		if route.RatePpm > group.BestRatePpm {
			group.BestRatePpm = route.RatePpm
		}
		if route.RatePpm < group.WorstRatePpm {
			group.WorstRatePpm = route.RatePpm
		}
	}
	out := make([]RouteGroup, 0, len(index))
	for _, group := range index {
		sort.Slice(group.RouteIDs, func(i, j int) bool {
			return group.RouteIDs[i] < group.RouteIDs[j]
		})
		out = append(out, *group)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Pair.Source != out[j].Pair.Source {
			return out[i].Pair.Source < out[j].Pair.Source
		}
		return out[i].Pair.Destination < out[j].Pair.Destination
	})
	return out
}

func PairStats(routes []domain.Route) []domain.PairStat {
	groups := GroupRoutes(routes)
	stats := make([]domain.PairStat, len(groups))
	for index, group := range groups {
		stats[index] = domain.PairStat{
			Pair:           group.Pair,
			Routes:         len(group.RouteIDs),
			ActiveRoutes:   group.Active,
			TotalExposure:  group.TotalExposure,
			TotalLiquidity: group.TotalOutput,
			BestRatePpm:    group.BestRatePpm,
			WorstRatePpm:   group.WorstRatePpm,
		}
	}
	return stats
}
