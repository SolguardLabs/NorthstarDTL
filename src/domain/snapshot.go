package domain

type RouteSnapshot struct {
	ID               RouteID     `json:"id"`
	SourceAsset      Asset       `json:"sourceAsset"`
	DestinationAsset Asset       `json:"destinationAsset"`
	RatePpm          int64       `json:"ratePpm"`
	CongestionBps    int64       `json:"congestionBps"`
	CongestionLevel  int64       `json:"congestionLevel"`
	LatencyMillis    int64       `json:"latencyMillis"`
	HealthBps        int64       `json:"healthBps"`
	Preferred        bool        `json:"preferred"`
	MaxExposure      Money       `json:"maxExposure"`
	Exposure         Money       `json:"exposure"`
	OutputLiquidity  Money       `json:"outputLiquidity"`
	Status           RouteStatus `json:"status"`
}

type QueueSnapshot struct {
	TicketID       TicketID     `json:"ticketId"`
	IntentID       IntentID     `json:"intentId"`
	PlannedRouteID RouteID      `json:"plannedRouteId"`
	ActiveRouteID  RouteID      `json:"activeRouteId"`
	Status         TicketStatus `json:"status"`
	ReadyEpoch     int64        `json:"readyEpoch"`
	QueueScore     int64        `json:"queueScore"`
	Amount         Money        `json:"amount"`
	Destination    Money        `json:"destinationAmount"`
	Fee            Money        `json:"fee"`
}

type Snapshot struct {
	Epoch       int64               `json:"epoch"`
	Balances    []Balance           `json:"balances"`
	Routes      []RouteSnapshot     `json:"routes"`
	Queue       []QueueSnapshot     `json:"queue"`
	Receipts    []SettlementReceipt `json:"receipts"`
	Events      []Event             `json:"events"`
	AuditIssues []AuditIssue        `json:"auditIssues"`
}

func NewRouteSnapshot(route Route) RouteSnapshot {
	return RouteSnapshot{
		ID:               route.ID,
		SourceAsset:      route.SourceAsset,
		DestinationAsset: route.DestinationAsset,
		RatePpm:          route.RatePpm,
		CongestionBps:    route.CongestionBps,
		CongestionLevel:  route.CongestionLevel,
		LatencyMillis:    route.LatencyMillis,
		HealthBps:        route.HealthBps,
		Preferred:        route.Preferred,
		MaxExposure:      route.MaxExposure,
		Exposure:         route.Exposure,
		OutputLiquidity:  route.OutputLiquidity,
		Status:           route.StatusOrDefault(),
	}
}

func NewQueueSnapshot(ticket Ticket) QueueSnapshot {
	return QueueSnapshot{
		TicketID:       ticket.ID,
		IntentID:       ticket.Intent.ID,
		PlannedRouteID: ticket.PlannedRouteID,
		ActiveRouteID:  ticket.ActiveRouteID,
		Status:         ticket.Status,
		ReadyEpoch:     ticket.ReadyEpoch,
		QueueScore:     ticket.QueueScore,
		Amount:         ticket.Intent.Amount,
		Destination:    ticket.Quote.NetOut,
		Fee:            ticket.Quote.RouteFee,
	}
}
