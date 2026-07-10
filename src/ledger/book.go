package ledger

import (
	"sort"

	"github.com/solguardlabs/northstardtl/src/domain"
)

type Book struct {
	balances map[accountKey]domain.Balance
	moves    []domain.LedgerMove
}

func NewBook(seeds []domain.BalanceSeed) (*Book, error) {
	book := &Book{balances: make(map[accountKey]domain.Balance)}
	for _, seed := range seeds {
		if err := book.Set(seed.Account, seed.Asset, seed.Available, seed.Reserved); err != nil {
			return nil, err
		}
	}
	return book, nil
}

func (book *Book) Clone() *Book {
	clone := &Book{
		balances: make(map[accountKey]domain.Balance, len(book.balances)),
		moves:    make([]domain.LedgerMove, len(book.moves)),
	}
	for key, balance := range book.balances {
		clone.balances[key] = balance
	}
	copy(clone.moves, book.moves)
	return clone
}

func (book *Book) Set(account domain.AccountID, asset domain.Asset, available, reserved domain.Money) error {
	balance := domain.Balance{Account: account, Asset: asset, Available: available, Reserved: reserved}
	if err := balance.Validate(); err != nil {
		return err
	}
	book.balances[key(account, asset)] = balance
	return nil
}

func (book *Book) Balance(account domain.AccountID, asset domain.Asset) domain.Balance {
	if found, ok := book.balances[key(account, asset)]; ok {
		return found
	}
	return domain.Balance{Account: account, Asset: asset}
}

func (book *Book) Credit(account domain.AccountID, asset domain.Asset, amount domain.Money, memo string) error {
	if amount < 0 {
		return domain.Invalid("credit amount cannot be negative")
	}
	if amount == 0 {
		return nil
	}
	balance := book.Balance(account, asset)
	balance.Available += amount
	book.balances[key(account, asset)] = balance
	book.moves = append(book.moves, domain.LedgerMove{ToAccount: account, Asset: asset, Amount: amount, Memo: memo})
	return nil
}

func (book *Book) Debit(account domain.AccountID, asset domain.Asset, amount domain.Money, memo string) error {
	if amount < 0 {
		return domain.Invalid("debit amount cannot be negative")
	}
	if amount == 0 {
		return nil
	}
	balance := book.Balance(account, asset)
	if balance.Available < amount {
		return domain.InsufficientFunds("account %s has %s %s available, needs %s", account, balance.Available, asset, amount)
	}
	balance.Available -= amount
	book.balances[key(account, asset)] = balance
	book.moves = append(book.moves, domain.LedgerMove{FromAccount: account, Asset: asset, Amount: amount, Memo: memo})
	return nil
}

func (book *Book) Transfer(from, to domain.AccountID, asset domain.Asset, amount domain.Money, memo string) error {
	if err := book.Debit(from, asset, amount, memo); err != nil {
		return err
	}
	if err := book.Credit(to, asset, amount, memo); err != nil {
		return err
	}
	book.moves = append(book.moves, domain.LedgerMove{FromAccount: from, ToAccount: to, Asset: asset, Amount: amount, Memo: memo})
	return nil
}

func (book *Book) Reserve(account domain.AccountID, asset domain.Asset, amount domain.Money, memo string) error {
	if amount < 0 {
		return domain.Invalid("reserve amount cannot be negative")
	}
	if amount == 0 {
		return nil
	}
	balance := book.Balance(account, asset)
	if balance.Available < amount {
		return domain.InsufficientFunds("account %s has %s %s available, needs reserve %s", account, balance.Available, asset, amount)
	}
	balance.Available -= amount
	balance.Reserved += amount
	book.balances[key(account, asset)] = balance
	book.moves = append(book.moves, domain.LedgerMove{FromAccount: account, Asset: asset, Amount: amount, Memo: memo})
	return nil
}

func (book *Book) Release(account domain.AccountID, asset domain.Asset, amount domain.Money, memo string) error {
	if amount < 0 {
		return domain.Invalid("release amount cannot be negative")
	}
	if amount == 0 {
		return nil
	}
	balance := book.Balance(account, asset)
	if balance.Reserved < amount {
		return domain.InsufficientFunds("account %s has %s %s reserved, needs release %s", account, balance.Reserved, asset, amount)
	}
	balance.Reserved -= amount
	balance.Available += amount
	book.balances[key(account, asset)] = balance
	book.moves = append(book.moves, domain.LedgerMove{ToAccount: account, Asset: asset, Amount: amount, Memo: memo})
	return nil
}

func (book *Book) SpendReserved(account domain.AccountID, asset domain.Asset, amount domain.Money, memo string) error {
	if amount < 0 {
		return domain.Invalid("spend reserved amount cannot be negative")
	}
	if amount == 0 {
		return nil
	}
	balance := book.Balance(account, asset)
	if balance.Reserved < amount {
		return domain.InsufficientFunds("account %s has %s %s reserved, needs spend %s", account, balance.Reserved, asset, amount)
	}
	balance.Reserved -= amount
	book.balances[key(account, asset)] = balance
	book.moves = append(book.moves, domain.LedgerMove{FromAccount: account, Asset: asset, Amount: amount, Memo: memo})
	return nil
}

func (book *Book) Snapshot() []domain.Balance {
	out := make([]domain.Balance, 0, len(book.balances))
	for _, balance := range book.balances {
		if balance.Available != 0 || balance.Reserved != 0 {
			out = append(out, balance)
		}
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Account != out[j].Account {
			return out[i].Account < out[j].Account
		}
		return out[i].Asset < out[j].Asset
	})
	return out
}

func (book *Book) Moves() []domain.LedgerMove {
	out := make([]domain.LedgerMove, len(book.moves))
	copy(out, book.moves)
	return out
}

func (book *Book) TotalsByAsset() map[domain.Asset]domain.Money {
	totals := make(map[domain.Asset]domain.Money)
	for _, balance := range book.balances {
		totals[balance.Asset] += balance.Total()
	}
	return totals
}

func (book *Book) HasAccount(account domain.AccountID, asset domain.Asset) bool {
	_, ok := book.balances[key(account, asset)]
	return ok
}
