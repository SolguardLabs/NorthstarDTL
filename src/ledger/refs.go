package ledger

import "github.com/solguardlabs/northstardtl/src/domain"

type accountKey struct {
	account domain.AccountID
	asset   domain.Asset
}

func key(account domain.AccountID, asset domain.Asset) accountKey {
	return accountKey{account: account, asset: asset}
}
