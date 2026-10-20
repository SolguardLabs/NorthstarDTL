package domain

import "testing"

func TestCapacityProjectionSeparatesSourceAndDestinationUnits(t *testing.T) {
	route := Route{
		ID: "route:test", SourceAsset: "usdc", DestinationAsset: "eurc",
		MaxExposure: 900_000, Exposure: 250_000, OutputLiquidity: 1_200_000,
		Status: RouteActive,
	}
	view := NewCapacityProjection(route)
	if view.SourceHeadroom != 650_000 || view.DestinationStressLiquidity != 1_020_000 {
		t.Fatalf("unexpected capacity view: %+v", view)
	}
	if view.DestinationStressLoss != 180_000 || !view.AdmissionOpen {
		t.Fatalf("unexpected stress or status: %+v", view)
	}
	route.Status = RoutePaused
	if NewCapacityProjection(route).AdmissionOpen {
		t.Fatal("paused route must not advertise open admission")
	}
}
