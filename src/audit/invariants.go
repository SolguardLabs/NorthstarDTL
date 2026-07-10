package audit

import (
	"fmt"

	"github.com/solguardlabs/northstardtl/src/domain"
	"github.com/solguardlabs/northstardtl/src/ledger"
)

type InvariantSet struct {
	RequireActiveLiquidity bool             `json:"requireActiveLiquidity"`
	MinimumHealthBps       int64            `json:"minimumHealthBps"`
	TrackedFeeAccount      domain.AccountID `json:"trackedFeeAccount"`
}

func DefaultInvariantSet() InvariantSet {
	return InvariantSet{
		RequireActiveLiquidity: true,
		MinimumHealthBps:       7_500,
		TrackedFeeAccount:      "acct:fees",
	}
}

func CheckInvariants(book *ledger.Book, routes []domain.Route, tickets []domain.Ticket, invariants InvariantSet) []domain.AuditIssue {
	issues := Check(book, routes, tickets)
	issues = append(issues, checkOperationalRoutes(routes, invariants)...)
	issues = append(issues, checkTicketReserves(book, tickets)...)
	return issues
}

func checkOperationalRoutes(routes []domain.Route, invariants InvariantSet) []domain.AuditIssue {
	issues := make([]domain.AuditIssue, 0)
	for _, route := range routes {
		if route.StatusOrDefault() != domain.RouteActive {
			continue
		}
		if route.HealthBps < invariants.MinimumHealthBps {
			issues = append(issues, domain.AuditIssue{
				Code:     "active_route_health_floor",
				Severity: domain.AuditWarning,
				RouteID:  route.ID,
				Message:  fmt.Sprintf("route %s is active below health floor", route.ID),
			})
		}
		if invariants.RequireActiveLiquidity && route.OutputLiquidity <= 0 {
			issues = append(issues, domain.AuditIssue{
				Code:     "active_route_empty_liquidity",
				Severity: domain.AuditWarning,
				RouteID:  route.ID,
				Message:  fmt.Sprintf("route %s has no output liquidity", route.ID),
			})
		}
	}
	return issues
}

func checkTicketReserves(book *ledger.Book, tickets []domain.Ticket) []domain.AuditIssue {
	required := make(map[reserveKey]domain.Money)
	for _, ticket := range tickets {
		if ticket.Status == domain.TicketExecuted || ticket.Status == domain.TicketRejected {
			continue
		}
		required[reserveKey{account: ticket.Intent.SourceAccount, asset: ticket.Intent.SourceAsset}] += ticket.Quote.ReserveAmount()
	}
	issues := make([]domain.AuditIssue, 0)
	for key, requiredAmount := range required {
		balance := book.Balance(key.account, key.asset)
		if balance.Reserved < requiredAmount {
			issues = append(issues, domain.AuditIssue{
				Code:     "pending_ticket_reserve_gap",
				Severity: domain.AuditCritical,
				Message:  fmt.Sprintf("account %s reserved %s below pending tickets %s", key.account, balance.Reserved, requiredAmount),
			})
		}
	}
	return issues
}

type reserveKey struct {
	account domain.AccountID
	asset   domain.Asset
}
