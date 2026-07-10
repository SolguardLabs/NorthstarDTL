package domain

type Balance struct {
	Account   AccountID `json:"account"`
	Asset     Asset     `json:"asset"`
	Available Money     `json:"available"`
	Reserved  Money     `json:"reserved"`
}

type BalanceSeed struct {
	Account   AccountID `json:"account"`
	Asset     Asset     `json:"asset"`
	Available Money     `json:"available"`
	Reserved  Money     `json:"reserved,omitempty"`
}

type LedgerMove struct {
	FromAccount AccountID `json:"fromAccount,omitempty"`
	ToAccount   AccountID `json:"toAccount,omitempty"`
	Asset       Asset     `json:"asset"`
	Amount      Money     `json:"amount"`
	Memo        string    `json:"memo,omitempty"`
}

func (b Balance) Total() Money {
	return b.Available + b.Reserved
}

func (b Balance) Validate() error {
	if !b.Account.Valid() || !b.Asset.Valid() {
		return Invalid("balance account and asset are required")
	}
	if b.Available < 0 || b.Reserved < 0 {
		return Invalid("balance cannot be negative")
	}
	return nil
}
