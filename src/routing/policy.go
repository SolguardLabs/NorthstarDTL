package routing

import "github.com/solguardlabs/northstardtl/src/domain"

type Policy struct {
	QuoteTTL               int64        `json:"quoteTtl"`
	MinHealthBps           int64        `json:"minHealthBps"`
	MaxAdmissionCongestion int64        `json:"maxAdmissionCongestion"`
	MaxFallbackCongestion  int64        `json:"maxFallbackCongestion"`
	PreferredRouteBonus    int64        `json:"preferredRouteBonus"`
	ClientPreferenceBonus  int64        `json:"clientPreferenceBonus"`
	HealthWeight           int64        `json:"healthWeight"`
	LatencyPenalty         int64        `json:"latencyPenalty"`
	ExposurePenaltyBps     int64        `json:"exposurePenaltyBps"`
	CongestionWeight       int64        `json:"congestionWeight"`
	MinimumNetOut          domain.Money `json:"minimumNetOut"`
}

func DefaultPolicy() Policy {
	return Policy{
		QuoteTTL:               4,
		MinHealthBps:           7_500,
		MaxAdmissionCongestion: 2_500,
		MaxFallbackCongestion:  4_500,
		PreferredRouteBonus:    40_000,
		ClientPreferenceBonus:  55_000,
		HealthWeight:           6,
		LatencyPenalty:         4,
		ExposurePenaltyBps:     800,
		CongestionWeight:       8,
		MinimumNetOut:          1,
	}
}
