package domain

type TicketStatus string

const (
	TicketQueued   TicketStatus = "queued"
	TicketReady    TicketStatus = "ready"
	TicketExecuted TicketStatus = "executed"
	TicketRejected TicketStatus = "rejected"
)

type Ticket struct {
	ID              TicketID     `json:"id"`
	Intent          Intent       `json:"intent"`
	PlannedRouteID  RouteID      `json:"plannedRouteId"`
	ActiveRouteID   RouteID      `json:"activeRouteId"`
	FallbackFrom    RouteID      `json:"fallbackFrom,omitempty"`
	Quote           Quote        `json:"quote"`
	Status          TicketStatus `json:"status"`
	CreatedEpoch    int64        `json:"createdEpoch"`
	ReadyEpoch      int64        `json:"readyEpoch"`
	Attempts        int64        `json:"attempts"`
	QueueScore      int64        `json:"queueScore"`
	LastTransition  int64        `json:"lastTransition"`
	AdmissionReason string       `json:"admissionReason,omitempty"`
}

type SubmitIntentResponse struct {
	Ticket Ticket `json:"ticket"`
	Quote  Quote  `json:"quote"`
}

func NewTicket(id TicketID, intent Intent, quote Quote, now int64) Ticket {
	priority := intent.PriorityOrDefault()
	ready := now + priority.ReadyDelay()
	if quote.ReadyEpoch > ready {
		ready = quote.ReadyEpoch
	}
	status := TicketQueued
	if ready <= now {
		status = TicketReady
	}
	return Ticket{
		ID:             id,
		Intent:         intent,
		PlannedRouteID: quote.RouteID,
		ActiveRouteID:  quote.RouteID,
		Quote:          quote,
		Status:         status,
		CreatedEpoch:   now,
		ReadyEpoch:     ready,
		QueueScore:     priority.QueueWeight() + quote.Score,
		LastTransition: now,
	}
}

func (t Ticket) IsReady(now int64) bool {
	return (t.Status == TicketReady || t.Status == TicketQueued) && t.ReadyEpoch <= now
}

func (t Ticket) MarkReady(now int64) Ticket {
	if t.Status == TicketQueued && t.ReadyEpoch <= now {
		t.Status = TicketReady
		t.LastTransition = now
	}
	return t
}

func (t Ticket) WithActiveRoute(route RouteID, now int64) Ticket {
	if route != "" && route != t.ActiveRouteID {
		if t.FallbackFrom == "" {
			t.FallbackFrom = t.ActiveRouteID
		}
		t.ActiveRouteID = route
		t.LastTransition = now
	}
	return t
}

func (t Ticket) MarkExecuted(now int64) Ticket {
	t.Status = TicketExecuted
	t.LastTransition = now
	return t
}

func (t Ticket) MarkRejected(now int64) Ticket {
	t.Status = TicketRejected
	t.LastTransition = now
	return t
}
