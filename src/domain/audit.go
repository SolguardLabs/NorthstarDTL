package domain

type AuditSeverity string

const (
	AuditInfo     AuditSeverity = "info"
	AuditWarning  AuditSeverity = "warning"
	AuditCritical AuditSeverity = "critical"
)

type AuditIssue struct {
	Code     string        `json:"code"`
	Severity AuditSeverity `json:"severity"`
	Message  string        `json:"message"`
	RouteID  RouteID       `json:"routeId,omitempty"`
	TicketID TicketID      `json:"ticketId,omitempty"`
}

func InfoIssue(code, message string) AuditIssue {
	return AuditIssue{Code: code, Severity: AuditInfo, Message: message}
}

func WarningIssue(code, message string) AuditIssue {
	return AuditIssue{Code: code, Severity: AuditWarning, Message: message}
}
