package scenario

import (
	"github.com/solguardlabs/northstardtl/src/api"
	"github.com/solguardlabs/northstardtl/src/domain"
	"github.com/solguardlabs/northstardtl/src/routing"
)

type BootstrapAlias struct {
	Epoch    int64                `json:"epoch"`
	Policy   routing.Policy       `json:"policy"`
	Routes   []domain.Route       `json:"routes"`
	Balances []domain.BalanceSeed `json:"balances"`
}

func FromAPI(bootstrap api.Bootstrap) BootstrapAlias {
	return BootstrapAlias{
		Epoch:    bootstrap.Epoch,
		Policy:   bootstrap.Policy,
		Routes:   bootstrap.Routes,
		Balances: bootstrap.Balances,
	}
}

func (alias BootstrapAlias) ToAPI() api.Bootstrap {
	return api.Bootstrap{
		Epoch:    alias.Epoch,
		Policy:   alias.Policy,
		Routes:   alias.Routes,
		Balances: alias.Balances,
	}
}
