package domain

type SettlementReceipt struct {
	ID                   ReceiptID `json:"id"`
	TicketID             TicketID  `json:"ticketId"`
	IntentID             IntentID  `json:"intentId"`
	RouteID              RouteID   `json:"routeId"`
	QuotedRouteID        RouteID   `json:"quotedRouteId"`
	SourceAccount        AccountID `json:"sourceAccount"`
	DestinationAccount   AccountID `json:"destinationAccount"`
	SourceAsset          Asset     `json:"sourceAsset"`
	DestinationAsset     Asset     `json:"destinationAsset"`
	AmountIn             Money     `json:"amountIn"`
	DestinationAmount    Money     `json:"destinationAmount"`
	RouteFee             Money     `json:"routeFee"`
	CongestionPenalty    Money     `json:"congestionPenalty"`
	ExecutedEpoch        int64     `json:"executedEpoch"`
	UsedFallback         bool      `json:"usedFallback"`
	FallbackFrom         RouteID   `json:"fallbackFrom,omitempty"`
	RouteExposureAfter   Money     `json:"routeExposureAfter"`
	RouteLiquidityAfter  Money     `json:"routeLiquidityAfter"`
	QueueAttempts        int64     `json:"queueAttempts"`
	SettlementLedgerMemo string    `json:"settlementLedgerMemo,omitempty"`
}

type ExecuteRequest struct {
	Count int `json:"count"`
}

type ExecuteResponse struct {
	Receipts []SettlementReceipt `json:"receipts"`
	Deferred []Ticket            `json:"deferred"`
}
