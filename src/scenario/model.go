package scenario

import (
	"github.com/solguardlabs/northstardtl/src/api"
	"github.com/solguardlabs/northstardtl/src/domain"
	"github.com/solguardlabs/northstardtl/src/routing"
)

type Definition struct {
	Name      string        `json:"name"`
	Bootstrap api.Bootstrap `json:"bootstrap"`
	Actions   []Action      `json:"actions"`
}

type Action struct {
	Type        string                  `json:"type"`
	Label       string                  `json:"label"`
	Intent      domain.Intent           `json:"intent,omitempty"`
	Count       int                     `json:"count,omitempty"`
	Delta       int64                   `json:"delta,omitempty"`
	RouteUpdate domain.RoutePatch       `json:"routeUpdate,omitempty"`
	Congestion  routing.CongestionDelta `json:"congestion,omitempty"`
	ExpectError domain.ErrorCode        `json:"expectError,omitempty"`
}

type Result struct {
	Name     string          `json:"name"`
	Results  []ActionResult  `json:"results"`
	Snapshot domain.Snapshot `json:"snapshot"`
}

type ActionResult struct {
	Type     string                       `json:"type"`
	Label    string                       `json:"label"`
	Quotes   []domain.Quote               `json:"quotes,omitempty"`
	Submit   *domain.SubmitIntentResponse `json:"submit,omitempty"`
	Execute  *domain.ExecuteResponse      `json:"execute,omitempty"`
	Epoch    *int64                       `json:"epoch,omitempty"`
	Route    *domain.Route                `json:"route,omitempty"`
	Snapshot *domain.Snapshot             `json:"snapshot,omitempty"`
	Error    *domain.DomainError          `json:"error,omitempty"`
}
