package domain

type RouteStatus string

const (
	RouteActive   RouteStatus = "active"
	RouteDraining RouteStatus = "draining"
	RoutePaused   RouteStatus = "paused"
)

type Route struct {
	ID                 RouteID     `json:"id"`
	DisplayName        string      `json:"displayName"`
	SourceAsset        Asset       `json:"sourceAsset"`
	DestinationAsset   Asset       `json:"destinationAsset"`
	SourceTreasury     AccountID   `json:"sourceTreasury"`
	DestinationVault   AccountID   `json:"destinationVault"`
	FeeAccount         AccountID   `json:"feeAccount"`
	RatePpm            int64       `json:"ratePpm"`
	BaseFeeBps         int64       `json:"baseFeeBps"`
	FlatFee            Money       `json:"flatFee"`
	CongestionBps      int64       `json:"congestionBps"`
	CongestionLevel    int64       `json:"congestionLevel"`
	LatencyMillis      int64       `json:"latencyMillis"`
	HealthBps          int64       `json:"healthBps"`
	Preferred          bool        `json:"preferred"`
	PreferredBias      int64       `json:"preferredBias"`
	MinAmount          Money       `json:"minAmount"`
	MaxAmount          Money       `json:"maxAmount"`
	MaxExposure        Money       `json:"maxExposure"`
	Exposure           Money       `json:"exposure"`
	OutputLiquidity    Money       `json:"outputLiquidity"`
	Status             RouteStatus `json:"status"`
	SettlementDelay    int64       `json:"settlementDelay"`
	LastCongestionSlot int64       `json:"lastCongestionSlot"`
}

type RoutePatch struct {
	ID                 RouteID      `json:"id"`
	Status             *RouteStatus `json:"status,omitempty"`
	RatePpm            *int64       `json:"ratePpm,omitempty"`
	BaseFeeBps         *int64       `json:"baseFeeBps,omitempty"`
	FlatFee            *Money       `json:"flatFee,omitempty"`
	CongestionBps      *int64       `json:"congestionBps,omitempty"`
	CongestionLevel    *int64       `json:"congestionLevel,omitempty"`
	LatencyMillis      *int64       `json:"latencyMillis,omitempty"`
	HealthBps          *int64       `json:"healthBps,omitempty"`
	Preferred          *bool        `json:"preferred,omitempty"`
	PreferredBias      *int64       `json:"preferredBias,omitempty"`
	MinAmount          *Money       `json:"minAmount,omitempty"`
	MaxAmount          *Money       `json:"maxAmount,omitempty"`
	MaxExposure        *Money       `json:"maxExposure,omitempty"`
	Exposure           *Money       `json:"exposure,omitempty"`
	OutputLiquidity    *Money       `json:"outputLiquidity,omitempty"`
	SettlementDelay    *int64       `json:"settlementDelay,omitempty"`
	LastCongestionSlot *int64       `json:"lastCongestionSlot,omitempty"`
}

func (r Route) Validate() error {
	if !r.ID.Valid() {
		return Invalid("route id is required")
	}
	if !r.SourceAsset.Valid() || !r.DestinationAsset.Valid() {
		return Invalid("route assets are required")
	}
	if r.SourceAsset == r.DestinationAsset {
		return Invalid("route asset pair must cross assets")
	}
	if !r.SourceTreasury.Valid() || !r.DestinationVault.Valid() || !r.FeeAccount.Valid() {
		return Invalid("route accounts are required")
	}
	if r.RatePpm <= 0 {
		return Invalid("route %s rate must be positive", r.ID)
	}
	if r.BaseFeeBps < 0 || r.FlatFee < 0 {
		return Invalid("route %s fee cannot be negative", r.ID)
	}
	if r.CongestionBps < 0 || r.CongestionLevel < 0 {
		return Invalid("route %s congestion cannot be negative", r.ID)
	}
	if r.HealthBps < 0 || r.HealthBps > BpsScale {
		return Invalid("route %s health must be within bps scale", r.ID)
	}
	if r.MaxAmount > 0 && r.MinAmount > r.MaxAmount {
		return Invalid("route %s min amount exceeds max amount", r.ID)
	}
	if r.MaxExposure < 0 || r.Exposure < 0 || r.OutputLiquidity < 0 {
		return Invalid("route %s liquidity fields cannot be negative", r.ID)
	}
	switch r.StatusOrDefault() {
	case RouteActive, RouteDraining, RoutePaused:
		return nil
	default:
		return Invalid("route %s has unsupported status %q", r.ID, r.Status)
	}
}

func (r Route) StatusOrDefault() RouteStatus {
	if r.Status == "" {
		return RouteActive
	}
	return r.Status
}

func (r Route) PairMatches(intent Intent) bool {
	return r.SourceAsset == intent.SourceAsset && r.DestinationAsset == intent.DestinationAsset
}

func (r Route) ActiveForAdmission() bool {
	return r.StatusOrDefault() == RouteActive
}

func (r Route) ActiveForFallback() bool {
	return r.StatusOrDefault() == RouteActive || r.StatusOrDefault() == RouteDraining
}

func (r Route) AmountAllowed(amount Money) bool {
	if r.MinAmount > 0 && amount < r.MinAmount {
		return false
	}
	if r.MaxAmount > 0 && amount > r.MaxAmount {
		return false
	}
	return true
}

func (r Route) RemainingExposure() Money {
	if r.MaxExposure <= r.Exposure {
		return 0
	}
	return r.MaxExposure - r.Exposure
}

func (r Route) CanCarry(intent Intent, destinationAmount Money) bool {
	return r.PairMatches(intent) &&
		r.AmountAllowed(intent.Amount) &&
		r.RemainingExposure() >= intent.Amount &&
		r.OutputLiquidity >= destinationAmount
}

func (r Route) ApplyPatch(patch RoutePatch) Route {
	if patch.Status != nil {
		r.Status = *patch.Status
	}
	if patch.RatePpm != nil {
		r.RatePpm = *patch.RatePpm
	}
	if patch.BaseFeeBps != nil {
		r.BaseFeeBps = *patch.BaseFeeBps
	}
	if patch.FlatFee != nil {
		r.FlatFee = *patch.FlatFee
	}
	if patch.CongestionBps != nil {
		r.CongestionBps = *patch.CongestionBps
	}
	if patch.CongestionLevel != nil {
		r.CongestionLevel = *patch.CongestionLevel
	}
	if patch.LatencyMillis != nil {
		r.LatencyMillis = *patch.LatencyMillis
	}
	if patch.HealthBps != nil {
		r.HealthBps = *patch.HealthBps
	}
	if patch.Preferred != nil {
		r.Preferred = *patch.Preferred
	}
	if patch.PreferredBias != nil {
		r.PreferredBias = *patch.PreferredBias
	}
	if patch.MinAmount != nil {
		r.MinAmount = *patch.MinAmount
	}
	if patch.MaxAmount != nil {
		r.MaxAmount = *patch.MaxAmount
	}
	if patch.MaxExposure != nil {
		r.MaxExposure = *patch.MaxExposure
	}
	if patch.Exposure != nil {
		r.Exposure = *patch.Exposure
	}
	if patch.OutputLiquidity != nil {
		r.OutputLiquidity = *patch.OutputLiquidity
	}
	if patch.SettlementDelay != nil {
		r.SettlementDelay = *patch.SettlementDelay
	}
	if patch.LastCongestionSlot != nil {
		r.LastCongestionSlot = *patch.LastCongestionSlot
	}
	return r
}
