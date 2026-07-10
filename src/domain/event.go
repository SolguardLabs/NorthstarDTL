package domain

type EventType string

const (
	EventRouteQuoted     EventType = "route_quoted"
	EventTicketSubmitted EventType = "ticket_submitted"
	EventRouteUpdated    EventType = "route_updated"
	EventEpochAdvanced   EventType = "epoch_advanced"
	EventTicketExecuted  EventType = "ticket_executed"
	EventTicketDeferred  EventType = "ticket_deferred"
	EventFallbackApplied EventType = "fallback_applied"
)

type Event struct {
	ID      EventID     `json:"id"`
	Type    EventType   `json:"type"`
	Epoch   int64       `json:"epoch"`
	RouteID RouteID     `json:"routeId,omitempty"`
	Ticket  TicketID    `json:"ticketId,omitempty"`
	Intent  IntentID    `json:"intentId,omitempty"`
	Message string      `json:"message,omitempty"`
	Data    StringTable `json:"data,omitempty"`
}

type EventSink struct {
	next   int64
	events []Event
}

func NewEventSink() *EventSink {
	return &EventSink{next: 1}
}

func (sink *EventSink) Append(event Event) Event {
	if event.ID == "" {
		event.ID = NewEventID(sink.next)
		sink.next++
	}
	if event.Data != nil {
		event.Data = event.Data.Clone()
	}
	sink.events = append(sink.events, event)
	return event
}

func (sink *EventSink) Snapshot() []Event {
	out := make([]Event, len(sink.events))
	copy(out, sink.events)
	return out
}
