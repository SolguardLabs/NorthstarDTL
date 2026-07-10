package domain

func KnownPriorities() []Priority {
	return []Priority{PriorityStandard, PriorityPreferred, PriorityUrgent}
}

func ParsePriority(value string) Priority {
	switch Priority(value) {
	case PriorityStandard, PriorityPreferred, PriorityUrgent:
		return Priority(value)
	default:
		return PriorityStandard
	}
}

func (p Priority) IsExpedited() bool {
	return p == PriorityPreferred || p == PriorityUrgent
}

func (p Priority) SettlementClass() string {
	switch p {
	case PriorityUrgent:
		return "immediate"
	case PriorityPreferred:
		return "fast"
	default:
		return "batch"
	}
}
