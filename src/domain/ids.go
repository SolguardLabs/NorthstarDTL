package domain

import (
	"fmt"
	"strings"
)

type AccountID string
type Asset string
type RouteID string
type IntentID string
type TicketID string
type ReceiptID string
type EventID string

func (id AccountID) Valid() bool { return strings.TrimSpace(string(id)) != "" }
func (id Asset) Valid() bool     { return strings.TrimSpace(string(id)) != "" }
func (id RouteID) Valid() bool   { return strings.TrimSpace(string(id)) != "" }
func (id IntentID) Valid() bool  { return strings.TrimSpace(string(id)) != "" }
func (id TicketID) Valid() bool  { return strings.TrimSpace(string(id)) != "" }
func (id ReceiptID) Valid() bool { return strings.TrimSpace(string(id)) != "" }
func (id EventID) Valid() bool   { return strings.TrimSpace(string(id)) != "" }

func NewTicketID(intent IntentID, sequence int64) TicketID {
	return TicketID(fmt.Sprintf("ticket:%s:%06d", sanitize(string(intent)), sequence))
}

func NewReceiptID(ticket TicketID, sequence int64) ReceiptID {
	return ReceiptID(fmt.Sprintf("receipt:%s:%06d", sanitize(string(ticket)), sequence))
}

func NewEventID(sequence int64) EventID {
	return EventID(fmt.Sprintf("event:%06d", sequence))
}

func sanitize(value string) string {
	value = strings.TrimSpace(value)
	value = strings.TrimPrefix(value, "intent:")
	value = strings.TrimPrefix(value, "ticket:")
	if value == "" {
		return "unknown"
	}
	replacer := strings.NewReplacer(":", "-", "/", "-", " ", "-")
	return replacer.Replace(value)
}
