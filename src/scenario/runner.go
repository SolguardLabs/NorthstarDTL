package scenario

import (
	"fmt"

	"github.com/solguardlabs/northstardtl/src/api"
	"github.com/solguardlabs/northstardtl/src/domain"
)

func Run(definition Definition) (Result, error) {
	bootstrap := definition.Bootstrap
	if len(bootstrap.Routes) == 0 {
		bootstrap = DefaultBootstrap()
	}
	service, err := api.NewService(bootstrap)
	if err != nil {
		return Result{}, err
	}
	results := make([]ActionResult, 0, len(definition.Actions))
	for _, action := range definition.Actions {
		result, err := runAction(service, action)
		if err != nil {
			if action.ExpectError != "" {
				if converted, ok := domain.AsDomainError(err); ok && converted.Code == action.ExpectError {
					result = ActionResult{
						Type:  action.Type,
						Label: action.Label,
						Error: &converted,
					}
					results = append(results, result)
					continue
				}
			}
			return Result{}, fmt.Errorf("action %q failed: %w", action.Type, err)
		}
		if action.ExpectError != "" {
			return Result{}, fmt.Errorf("action %q expected error %s", action.Type, action.ExpectError)
		}
		results = append(results, result)
	}
	return Result{
		Name:     definition.Name,
		Results:  results,
		Snapshot: service.Snapshot(),
	}, nil
}

func runAction(service *api.Service, action Action) (ActionResult, error) {
	switch action.Type {
	case "quote":
		book, err := service.QuoteIntent(action.Intent)
		if err != nil {
			return ActionResult{}, err
		}
		return ActionResult{Type: action.Type, Label: action.Label, Quotes: book.Quotes}, nil
	case "submit":
		response, err := service.SubmitIntent(domain.SubmitIntentRequest{Intent: action.Intent})
		if err != nil {
			return ActionResult{}, err
		}
		return ActionResult{Type: action.Type, Label: action.Label, Submit: &response}, nil
	case "execute":
		response, err := service.Execute(domain.ExecuteRequest{Count: action.Count})
		if err != nil {
			return ActionResult{}, err
		}
		return ActionResult{Type: action.Type, Label: action.Label, Execute: &response}, nil
	case "advance_epoch":
		epoch := service.AdvanceEpoch(action.Delta)
		return ActionResult{Type: action.Type, Label: action.Label, Epoch: &epoch}, nil
	case "update_route":
		route, err := service.UpdateRoute(action.RouteUpdate)
		if err != nil {
			return ActionResult{}, err
		}
		return ActionResult{Type: action.Type, Label: action.Label, Route: &route}, nil
	case "congestion":
		route, err := service.ApplyCongestion(action.Congestion)
		if err != nil {
			return ActionResult{}, err
		}
		return ActionResult{Type: action.Type, Label: action.Label, Route: &route}, nil
	case "snapshot":
		snapshot := service.Snapshot()
		return ActionResult{Type: action.Type, Label: action.Label, Snapshot: &snapshot}, nil
	default:
		return ActionResult{}, domain.NewError(domain.ErrUnknownAction, "unknown action type %q", action.Type)
	}
}
