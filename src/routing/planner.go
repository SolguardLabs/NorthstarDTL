package routing

import (
	"sort"

	"github.com/solguardlabs/northstardtl/src/domain"
)

type PlanMode string

const (
	PlanModeAdmission PlanMode = "admission"
	PlanModeFallback  PlanMode = "fallback"
	PlanModeObserve   PlanMode = "observe"
)

type PlanStep struct {
	Rank              int                `json:"rank"`
	RouteID           domain.RouteID     `json:"routeId"`
	Mode              PlanMode           `json:"mode"`
	Quote             domain.Quote       `json:"quote"`
	Status            domain.RouteStatus `json:"status"`
	RemainingExposure domain.Money       `json:"remainingExposure"`
	OutputLiquidity   domain.Money       `json:"outputLiquidity"`
	Admissible        bool               `json:"admissible"`
	Explanation       string             `json:"explanation"`
}

type RoutePlan struct {
	IntentID     domain.IntentID `json:"intentId"`
	Pair         domain.PairKey  `json:"pair"`
	CreatedEpoch int64           `json:"createdEpoch"`
	Mode         PlanMode        `json:"mode"`
	Steps        []PlanStep      `json:"steps"`
	Selected     *PlanStep       `json:"selected,omitempty"`
}

func BuildPlan(intent domain.Intent, routes []domain.Route, scorer Scorer, now int64, mode PlanMode) RoutePlan {
	steps := make([]PlanStep, 0, len(routes))
	for _, route := range routes {
		if !route.PairMatches(intent) {
			continue
		}
		quote := scorer.Quote(route, intent, now)
		step := PlanStep{
			RouteID:           route.ID,
			Mode:              mode,
			Quote:             quote,
			Status:            route.StatusOrDefault(),
			RemainingExposure: route.RemainingExposure(),
			OutputLiquidity:   route.OutputLiquidity,
			Explanation:       quote.Reason,
		}
		switch mode {
		case PlanModeFallback:
			step.Admissible = scorer.FallbackAdmissible(route, intent, quote.NetOut)
		case PlanModeObserve:
			step.Admissible = route.PairMatches(intent)
		default:
			step.Admissible = scorer.Admissible(route, intent, quote)
		}
		steps = append(steps, step)
	}
	sort.SliceStable(steps, func(i, j int) bool {
		if steps[i].Quote.Score != steps[j].Quote.Score {
			return steps[i].Quote.Score > steps[j].Quote.Score
		}
		return steps[i].RouteID < steps[j].RouteID
	})
	for index := range steps {
		steps[index].Rank = index + 1
	}
	var selected *PlanStep
	for index := range steps {
		if steps[index].Admissible {
			copy := steps[index]
			selected = &copy
			break
		}
	}
	return RoutePlan{
		IntentID:     intent.ID,
		Pair:         domain.PairKey{Source: intent.SourceAsset, Destination: intent.DestinationAsset},
		CreatedEpoch: now,
		Mode:         mode,
		Steps:        steps,
		Selected:     selected,
	}
}

func (plan RoutePlan) HasSelection() bool {
	return plan.Selected != nil
}

func (plan RoutePlan) TopRoutes(limit int) []PlanStep {
	if limit <= 0 || limit > len(plan.Steps) {
		limit = len(plan.Steps)
	}
	out := make([]PlanStep, limit)
	copy(out, plan.Steps[:limit])
	return out
}

func (plan RoutePlan) Rejected() []PlanStep {
	out := make([]PlanStep, 0)
	for _, step := range plan.Steps {
		if !step.Admissible {
			out = append(out, step)
		}
	}
	return out
}

func (plan RoutePlan) SelectedRouteID() domain.RouteID {
	if plan.Selected == nil {
		return ""
	}
	return plan.Selected.RouteID
}
