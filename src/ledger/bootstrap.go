package ledger

import "github.com/solguardlabs/northstardtl/src/domain"

func RouteSeeds(routes []domain.Route) []domain.BalanceSeed {
	seeds := make([]domain.BalanceSeed, 0, len(routes)*3)
	for _, route := range routes {
		seeds = append(seeds,
			domain.BalanceSeed{Account: route.SourceTreasury, Asset: route.SourceAsset, Available: 0},
			domain.BalanceSeed{Account: route.DestinationVault, Asset: route.DestinationAsset, Available: route.OutputLiquidity},
			domain.BalanceSeed{Account: route.FeeAccount, Asset: route.SourceAsset, Available: 0},
		)
	}
	return seeds
}
