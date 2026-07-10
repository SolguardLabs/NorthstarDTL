package domain

type PairKey struct {
	Source      Asset `json:"source"`
	Destination Asset `json:"destination"`
}

type RouteMetrics struct {
	RouteID             RouteID     `json:"routeId"`
	Pair                PairKey     `json:"pair"`
	Status              RouteStatus `json:"status"`
	RatePpm             int64       `json:"ratePpm"`
	HealthBps           int64       `json:"healthBps"`
	CongestionBps       int64       `json:"congestionBps"`
	CongestionLevel     int64       `json:"congestionLevel"`
	LatencyMillis       int64       `json:"latencyMillis"`
	Exposure            Money       `json:"exposure"`
	MaxExposure         Money       `json:"maxExposure"`
	OutputLiquidity     Money       `json:"outputLiquidity"`
	ExposureUtilization int64       `json:"exposureUtilization"`
	LiquidityWeight     int64       `json:"liquidityWeight"`
	OperationalScore    int64       `json:"operationalScore"`
}

type RouterStats struct {
	Epoch             int64          `json:"epoch"`
	RouteCount        int            `json:"routeCount"`
	ActiveRoutes      int            `json:"activeRoutes"`
	DrainingRoutes    int            `json:"drainingRoutes"`
	PausedRoutes      int            `json:"pausedRoutes"`
	TotalExposure     Money          `json:"totalExposure"`
	TotalLiquidity    Money          `json:"totalLiquidity"`
	WeightedHealthBps int64          `json:"weightedHealthBps"`
	ByPair            []PairStat     `json:"byPair"`
	Routes            []RouteMetrics `json:"routes"`
}

type PairStat struct {
	Pair           PairKey `json:"pair"`
	Routes         int     `json:"routes"`
	ActiveRoutes   int     `json:"activeRoutes"`
	TotalExposure  Money   `json:"totalExposure"`
	TotalLiquidity Money   `json:"totalLiquidity"`
	BestRatePpm    int64   `json:"bestRatePpm"`
	WorstRatePpm   int64   `json:"worstRatePpm"`
}

type SettlementStats struct {
	ReceiptCount       int   `json:"receiptCount"`
	FallbackCount      int   `json:"fallbackCount"`
	TotalSourceAmount  Money `json:"totalSourceAmount"`
	TotalDestAmount    Money `json:"totalDestinationAmount"`
	TotalFees          Money `json:"totalFees"`
	TotalCongestionCut Money `json:"totalCongestionCut"`
}

func NewRouteMetrics(route Route) RouteMetrics {
	utilization := int64(0)
	if route.MaxExposure > 0 {
		utilization = int64(route.Exposure) * BpsScale / int64(route.MaxExposure)
	}
	liquidityWeight := int64(route.OutputLiquidity / 1_000)
	operational := route.HealthBps*4 - route.CongestionBps*2 - route.CongestionLevel - route.LatencyMillis
	if route.Preferred {
		operational += route.PreferredBias / 10
	}
	return RouteMetrics{
		RouteID:             route.ID,
		Pair:                PairKey{Source: route.SourceAsset, Destination: route.DestinationAsset},
		Status:              route.StatusOrDefault(),
		RatePpm:             route.RatePpm,
		HealthBps:           route.HealthBps,
		CongestionBps:       route.CongestionBps,
		CongestionLevel:     route.CongestionLevel,
		LatencyMillis:       route.LatencyMillis,
		Exposure:            route.Exposure,
		MaxExposure:         route.MaxExposure,
		OutputLiquidity:     route.OutputLiquidity,
		ExposureUtilization: utilization,
		LiquidityWeight:     liquidityWeight,
		OperationalScore:    operational,
	}
}

func SummarizeReceipts(receipts []SettlementReceipt) SettlementStats {
	stats := SettlementStats{ReceiptCount: len(receipts)}
	for _, receipt := range receipts {
		if receipt.UsedFallback {
			stats.FallbackCount++
		}
		stats.TotalSourceAmount += receipt.AmountIn
		stats.TotalDestAmount += receipt.DestinationAmount
		stats.TotalFees += receipt.RouteFee
		stats.TotalCongestionCut += receipt.CongestionPenalty
	}
	return stats
}

func PairOf(route Route) PairKey {
	return PairKey{Source: route.SourceAsset, Destination: route.DestinationAsset}
}
