package domain

type Quote struct {
	RouteID           RouteID `json:"routeId"`
	AmountIn          Money   `json:"amountIn"`
	GrossOut          Money   `json:"grossOut"`
	CongestionPenalty Money   `json:"congestionPenalty"`
	RouteFee          Money   `json:"routeFee"`
	NetOut            Money   `json:"netOut"`
	Score             int64   `json:"score"`
	ReadyEpoch        int64   `json:"readyEpoch"`
	ExpiresEpoch      int64   `json:"expiresEpoch"`
	Reason            string  `json:"reason,omitempty"`
}

type QuoteBook struct {
	IntentID IntentID `json:"intentId"`
	Quotes   []Quote  `json:"quotes"`
}

func (q Quote) ReserveAmount() Money {
	return q.AmountIn + q.RouteFee
}

func (q Quote) MeetsIntent(intent Intent) bool {
	if q.NetOut <= 0 {
		return false
	}
	if intent.MinDestinationAmount > 0 && q.NetOut < intent.MinDestinationAmount {
		return false
	}
	if intent.MaxFee > 0 && q.RouteFee > intent.MaxFee {
		return false
	}
	return true
}

func (q Quote) BetterThan(other Quote) bool {
	if q.Score != other.Score {
		return q.Score > other.Score
	}
	if q.NetOut != other.NetOut {
		return q.NetOut > other.NetOut
	}
	if q.RouteFee != other.RouteFee {
		return q.RouteFee < other.RouteFee
	}
	return q.RouteID < other.RouteID
}
