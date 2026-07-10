package routing

import (
	"fmt"

	"github.com/solguardlabs/northstardtl/src/domain"
)

type Scorer struct {
	policy Policy
}

func NewScorer(policy Policy) Scorer {
	return Scorer{policy: policy}
}

func (scorer Scorer) Policy() Policy {
	return scorer.policy
}

func (scorer Scorer) Quote(route domain.Route, intent domain.Intent, now int64) domain.Quote {
	gross := intent.Amount.MulRate(route.RatePpm)
	congestion := scorer.congestionPenalty(route, gross, intent.Amount)
	fee := intent.Amount.MulBps(route.BaseFeeBps) + route.FlatFee
	net := gross - congestion
	if net < 0 {
		net = 0
	}
	quote := domain.Quote{
		RouteID:           route.ID,
		AmountIn:          intent.Amount,
		GrossOut:          gross,
		CongestionPenalty: congestion,
		RouteFee:          fee,
		NetOut:            net,
		ReadyEpoch:        now + route.SettlementDelay,
		ExpiresEpoch:      now + scorer.policy.QuoteTTL,
	}
	quote.Score = scorer.score(route, intent, quote)
	quote.Reason = scorer.reason(route, intent, quote)
	return quote
}

func (scorer Scorer) congestionPenalty(route domain.Route, gross domain.Money, amount domain.Money) domain.Money {
	pricePenalty := gross.MulBps(route.CongestionBps)
	queuePenalty := amount.MulBps(route.CongestionLevel)
	return pricePenalty + queuePenalty
}

func (scorer Scorer) score(route domain.Route, intent domain.Intent, quote domain.Quote) int64 {
	score := quote.NetOut.Int64() * 10
	score -= quote.RouteFee.Int64() * 3
	score += (route.HealthBps - scorer.policy.MinHealthBps) * scorer.policy.HealthWeight
	score -= route.LatencyMillis * scorer.policy.LatencyPenalty
	score -= (route.CongestionBps + route.CongestionLevel) * scorer.policy.CongestionWeight
	if route.Preferred {
		score += scorer.policy.PreferredRouteBonus + route.PreferredBias
	}
	if intent.Prefers(route.ID) {
		score += scorer.policy.ClientPreferenceBonus
	}
	if route.MaxExposure > 0 {
		usedBps := int64(route.Exposure) * domain.BpsScale / int64(route.MaxExposure)
		score -= usedBps * scorer.policy.ExposurePenaltyBps / domain.BpsScale
	}
	return score
}

func (scorer Scorer) reason(route domain.Route, intent domain.Intent, quote domain.Quote) string {
	if !route.PairMatches(intent) {
		return "asset-pair"
	}
	if !route.AmountAllowed(intent.Amount) {
		return "amount-window"
	}
	if route.HealthBps < scorer.policy.MinHealthBps {
		return "health"
	}
	if route.CongestionBps > scorer.policy.MaxAdmissionCongestion {
		return "congestion"
	}
	if route.RemainingExposure() < intent.Amount {
		return "exposure"
	}
	if quote.NetOut < scorer.policy.MinimumNetOut {
		return "minimum-output"
	}
	if !quote.MeetsIntent(intent) {
		return "intent-limits"
	}
	return fmt.Sprintf("score:%d", quote.Score)
}

func (scorer Scorer) Admissible(route domain.Route, intent domain.Intent, quote domain.Quote) bool {
	if !route.ActiveForAdmission() {
		return false
	}
	if route.HealthBps < scorer.policy.MinHealthBps {
		return false
	}
	if route.CongestionBps > scorer.policy.MaxAdmissionCongestion {
		return false
	}
	return route.CanCarry(intent, quote.NetOut) && quote.MeetsIntent(intent)
}

func (scorer Scorer) FallbackAdmissible(route domain.Route, intent domain.Intent, destination domain.Money) bool {
	if !route.ActiveForFallback() {
		return false
	}
	if route.HealthBps < scorer.policy.MinHealthBps {
		return false
	}
	if route.CongestionBps > scorer.policy.MaxFallbackCongestion {
		return false
	}
	return route.CanCarry(intent, destination)
}
