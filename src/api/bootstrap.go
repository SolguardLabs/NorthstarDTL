package api

import (
	"github.com/solguardlabs/northstardtl/src/domain"
	"github.com/solguardlabs/northstardtl/src/routing"
)

type Bootstrap struct {
	Epoch    int64                `json:"epoch"`
	Policy   routing.Policy       `json:"policy"`
	Routes   []domain.Route       `json:"routes"`
	Balances []domain.BalanceSeed `json:"balances"`
}

func (bootstrap Bootstrap) WithDefaults() Bootstrap {
	if bootstrap.Policy.QuoteTTL == 0 {
		bootstrap.Policy = routing.DefaultPolicy()
	}
	return bootstrap
}
