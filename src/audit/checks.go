package audit

import (
	"fmt"

	"github.com/solguardlabs/northstardtl/src/domain"
	"github.com/solguardlabs/northstardtl/src/ledger"
)

func Check(book *ledger.Book, routes []domain.Route, tickets []domain.Ticket) []domain.AuditIssue {
	issues := make([]domain.AuditIssue, 0)
	issues = append(issues, checkBalances(book)...)
	issues = append(issues, checkRoutes(routes)...)
	issues = append(issues, checkQueue(tickets)...)
	return issues
}

func checkBalances(book *ledger.Book) []domain.AuditIssue {
	issues := make([]domain.AuditIssue, 0)
	for _, balance := range book.Snapshot() {
		if balance.Available < 0 || balance.Reserved < 0 {
			issues = append(issues, domain.AuditIssue{
				Code:     "negative_balance",
				Severity: domain.AuditCritical,
				Message:  fmt.Sprintf("account %s has negative %s balance", balance.Account, balance.Asset),
			})
		}
	}
	return issues
}

func checkRoutes(routes []domain.Route) []domain.AuditIssue {
	issues := make([]domain.AuditIssue, 0)
	for _, route := range routes {
		if route.MaxExposure > 0 && route.Exposure > route.MaxExposure {
			issues = append(issues, domain.AuditIssue{
				Code:     "route_exposure_limit",
				Severity: domain.AuditCritical,
				Message:  fmt.Sprintf("route %s exposure exceeds configured limit", route.ID),
				RouteID:  route.ID,
			})
		}
		if route.OutputLiquidity < 0 {
			issues = append(issues, domain.AuditIssue{
				Code:     "route_liquidity_negative",
				Severity: domain.AuditCritical,
				Message:  fmt.Sprintf("route %s output liquidity is negative", route.ID),
				RouteID:  route.ID,
			})
		}
		if route.HealthBps < 5_000 && route.StatusOrDefault() == domain.RouteActive {
			issues = append(issues, domain.AuditIssue{
				Code:     "low_route_health",
				Severity: domain.AuditWarning,
				Message:  fmt.Sprintf("route %s remains active below operational health threshold", route.ID),
				RouteID:  route.ID,
			})
		}
	}
	return issues
}

func checkQueue(tickets []domain.Ticket) []domain.AuditIssue {
	issues := make([]domain.AuditIssue, 0)
	for _, ticket := range tickets {
		if ticket.Status == domain.TicketQueued && ticket.ReadyEpoch < ticket.CreatedEpoch {
			issues = append(issues, domain.AuditIssue{
				Code:     "queue_epoch_order",
				Severity: domain.AuditWarning,
				Message:  fmt.Sprintf("ticket %s has inconsistent queue epochs", ticket.ID),
				TicketID: ticket.ID,
			})
		}
		if ticket.Quote.NetOut <= 0 && ticket.Status != domain.TicketRejected {
			issues = append(issues, domain.AuditIssue{
				Code:     "empty_destination_quote",
				Severity: domain.AuditWarning,
				Message:  fmt.Sprintf("ticket %s has zero destination quote", ticket.ID),
				TicketID: ticket.ID,
			})
		}
	}
	return issues
}
