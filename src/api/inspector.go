package api

import (
	"sort"

	"github.com/solguardlabs/northstardtl/src/domain"
	"github.com/solguardlabs/northstardtl/src/routing"
	"github.com/solguardlabs/northstardtl/src/settlement"
)

type RouteBook struct {
	Epoch     int64                 `json:"epoch"`
	Policy    routing.Policy        `json:"policy"`
	Groups    []routing.RouteGroup  `json:"groups"`
	Metrics   []domain.RouteMetrics `json:"metrics"`
	PairStats []domain.PairStat     `json:"pairStats"`
}

type AccountView struct {
	Account   domain.AccountID `json:"account"`
	Balances  []domain.Balance `json:"balances"`
	Reserved  domain.Money     `json:"reserved"`
	Available domain.Money     `json:"available"`
}

type OperationsView struct {
	Epoch       int64                      `json:"epoch"`
	Queue       []domain.QueueSnapshot     `json:"queue"`
	Receipts    []domain.SettlementReceipt `json:"receipts"`
	Journal     settlement.Journal         `json:"journal"`
	Stats       domain.SettlementStats     `json:"stats"`
	AuditIssues []domain.AuditIssue        `json:"auditIssues"`
}

func (service *Service) RouteBook() RouteBook {
	routes := service.catalog.Routes()
	metrics := make([]domain.RouteMetrics, len(routes))
	for index, route := range routes {
		metrics[index] = domain.NewRouteMetrics(route)
	}
	sort.Slice(metrics, func(i, j int) bool {
		if metrics[i].OperationalScore != metrics[j].OperationalScore {
			return metrics[i].OperationalScore > metrics[j].OperationalScore
		}
		return metrics[i].RouteID < metrics[j].RouteID
	})
	return RouteBook{
		Epoch:     service.epoch,
		Policy:    service.Policy(),
		Groups:    routing.GroupRoutes(routes),
		Metrics:   metrics,
		PairStats: routing.PairStats(routes),
	}
}

func (service *Service) Account(account domain.AccountID) AccountView {
	balances := make([]domain.Balance, 0)
	var available domain.Money
	var reserved domain.Money
	for _, balance := range service.book.Snapshot() {
		if balance.Account != account {
			continue
		}
		balances = append(balances, balance)
		available += balance.Available
		reserved += balance.Reserved
	}
	return AccountView{
		Account:   account,
		Balances:  balances,
		Available: available,
		Reserved:  reserved,
	}
}

func (service *Service) Operations() OperationsView {
	receipts := service.engine.Receipts()
	journal := settlement.BuildJournal(service.engine.Tickets(), receipts)
	return OperationsView{
		Epoch:       service.epoch,
		Queue:       service.engine.QueueSnapshot(),
		Receipts:    receipts,
		Journal:     journal,
		Stats:       domain.SummarizeReceipts(receipts),
		AuditIssues: service.Snapshot().AuditIssues,
	}
}

func (service *Service) Explain(intent domain.Intent, mode routing.PlanMode) (routing.RoutePlan, error) {
	if err := intent.Validate(); err != nil {
		return routing.RoutePlan{}, err
	}
	return routing.BuildPlan(intent, service.catalog.Matching(intent), service.scorer, service.epoch, mode), nil
}
