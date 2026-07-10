package routing

import "github.com/solguardlabs/northstardtl/src/domain"

type CongestionDelta struct {
	RouteID          domain.RouteID `json:"routeId"`
	CongestionBps    int64          `json:"congestionBps"`
	CongestionLevel  int64          `json:"congestionLevel"`
	LastObservedSlot int64          `json:"lastObservedSlot"`
}

func ApplyCongestion(catalog *Catalog, delta CongestionDelta) (domain.Route, error) {
	patch := domain.RoutePatch{
		ID:                 delta.RouteID,
		CongestionBps:      &delta.CongestionBps,
		CongestionLevel:    &delta.CongestionLevel,
		LastCongestionSlot: &delta.LastObservedSlot,
	}
	return catalog.ApplyPatch(patch)
}

func SmoothCongestion(previous int64, observed int64) int64 {
	if previous <= 0 {
		return observed
	}
	if observed <= 0 {
		return previous / 2
	}
	return (previous*3 + observed) / 4
}
