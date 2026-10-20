package api_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/solguardlabs/northstardtl/src/api"
	"github.com/solguardlabs/northstardtl/src/domain"
	"github.com/solguardlabs/northstardtl/src/scenario"
)

func TestCapacityEndpointReportsRouteStress(t *testing.T) {
	service, err := api.NewService(scenario.DefaultBootstrap())
	if err != nil {
		t.Fatal(err)
	}
	request := httptest.NewRequest(http.MethodGet, "/capacity", nil)
	response := httptest.NewRecorder()
	api.NewHTTPHandler(service).ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("unexpected response status: %d", response.Code)
	}
	var views []domain.CapacityProjection
	if err := json.Unmarshal(response.Body.Bytes(), &views); err != nil {
		t.Fatal(err)
	}
	if len(views) != 4 {
		t.Fatalf("expected four route views, got %d", len(views))
	}
	for _, view := range views {
		if view.RouteID == "route:aurora-rfq" {
			if view.SourceHeadroom != 650_000 || view.DestinationStressLiquidity != 1_020_000 {
				t.Fatalf("unexpected aurora capacity: %+v", view)
			}
			return
		}
	}
	t.Fatal("aurora route missing")
}
