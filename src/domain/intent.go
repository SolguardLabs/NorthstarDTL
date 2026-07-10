package domain

type Priority string

const (
	PriorityStandard  Priority = "standard"
	PriorityPreferred Priority = "preferred"
	PriorityUrgent    Priority = "urgent"
)

type Intent struct {
	ID                   IntentID    `json:"id"`
	SourceAccount        AccountID   `json:"sourceAccount"`
	DestinationAccount   AccountID   `json:"destinationAccount"`
	SourceAsset          Asset       `json:"sourceAsset"`
	DestinationAsset     Asset       `json:"destinationAsset"`
	Amount               Money       `json:"amount"`
	MinDestinationAmount Money       `json:"minDestinationAmount,omitempty"`
	MaxFee               Money       `json:"maxFee"`
	Priority             Priority    `json:"priority"`
	PreferredRoutes      []RouteID   `json:"preferredRoutes,omitempty"`
	ClientTag            string      `json:"clientTag,omitempty"`
	Metadata             StringTable `json:"metadata,omitempty"`
}

type SubmitIntentRequest struct {
	Intent Intent `json:"intent"`
}

func (i Intent) Validate() error {
	if !i.ID.Valid() {
		return Invalid("intent id is required")
	}
	if !i.SourceAccount.Valid() || !i.DestinationAccount.Valid() {
		return Invalid("source and destination accounts are required")
	}
	if !i.SourceAsset.Valid() || !i.DestinationAsset.Valid() {
		return Invalid("source and destination assets are required")
	}
	if i.SourceAsset == i.DestinationAsset {
		return Invalid("asset pair must cross assets")
	}
	if !i.Amount.Positive() {
		return Invalid("amount must be positive")
	}
	if i.MaxFee < 0 {
		return Invalid("max fee cannot be negative")
	}
	switch i.PriorityOrDefault() {
	case PriorityStandard, PriorityPreferred, PriorityUrgent:
		return nil
	default:
		return Invalid("unsupported priority %q", i.Priority)
	}
}

func (i Intent) PriorityOrDefault() Priority {
	if i.Priority == "" {
		return PriorityStandard
	}
	return i.Priority
}

func (i Intent) Prefers(route RouteID) bool {
	for _, candidate := range i.PreferredRoutes {
		if candidate == route {
			return true
		}
	}
	return false
}

func (i Intent) RequiredReserve(fee Money) Money {
	return i.Amount + fee
}

func (p Priority) ReadyDelay() int64 {
	switch p {
	case PriorityUrgent:
		return 0
	case PriorityPreferred:
		return 1
	default:
		return 2
	}
}

func (p Priority) QueueWeight() int64 {
	switch p {
	case PriorityUrgent:
		return 30_000
	case PriorityPreferred:
		return 15_000
	default:
		return 5_000
	}
}
