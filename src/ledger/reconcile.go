package ledger

import (
	"fmt"
	"sort"

	"github.com/solguardlabs/northstardtl/src/domain"
)

type AssetTotal struct {
	Asset     domain.Asset `json:"asset"`
	Available domain.Money `json:"available"`
	Reserved  domain.Money `json:"reserved"`
	Total     domain.Money `json:"total"`
	Accounts  int          `json:"accounts"`
}

type ReconciliationIssue struct {
	Code     string           `json:"code"`
	Asset    domain.Asset     `json:"asset,omitempty"`
	Account  domain.AccountID `json:"account,omitempty"`
	Expected domain.Money     `json:"expected,omitempty"`
	Actual   domain.Money     `json:"actual,omitempty"`
	Message  string           `json:"message"`
}

func (book *Book) AssetTotals() []AssetTotal {
	type mutable struct {
		available domain.Money
		reserved  domain.Money
		accounts  int
	}
	byAsset := make(map[domain.Asset]mutable)
	for _, balance := range book.balances {
		next := byAsset[balance.Asset]
		next.available += balance.Available
		next.reserved += balance.Reserved
		next.accounts++
		byAsset[balance.Asset] = next
	}
	out := make([]AssetTotal, 0, len(byAsset))
	for asset, total := range byAsset {
		out = append(out, AssetTotal{
			Asset:     asset,
			Available: total.available,
			Reserved:  total.reserved,
			Total:     total.available + total.reserved,
			Accounts:  total.accounts,
		})
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].Asset < out[j].Asset
	})
	return out
}

func (book *Book) ReconcileExpected(expected map[domain.Asset]domain.Money) []ReconciliationIssue {
	issues := make([]ReconciliationIssue, 0)
	actual := book.TotalsByAsset()
	for asset, expectedTotal := range expected {
		actualTotal := actual[asset]
		if actualTotal != expectedTotal {
			issues = append(issues, ReconciliationIssue{
				Code:     "asset_total_mismatch",
				Asset:    asset,
				Expected: expectedTotal,
				Actual:   actualTotal,
				Message:  fmt.Sprintf("asset %s total mismatch", asset),
			})
		}
	}
	for asset, actualTotal := range actual {
		if _, ok := expected[asset]; !ok && actualTotal != 0 {
			issues = append(issues, ReconciliationIssue{
				Code:    "unexpected_asset_total",
				Asset:   asset,
				Actual:  actualTotal,
				Message: fmt.Sprintf("asset %s was not present in expected totals", asset),
			})
		}
	}
	return issues
}

func (book *Book) ReservedByAccount(account domain.AccountID) []domain.Balance {
	out := make([]domain.Balance, 0)
	for _, balance := range book.balances {
		if balance.Account == account && balance.Reserved > 0 {
			out = append(out, balance)
		}
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].Asset < out[j].Asset
	})
	return out
}

func (book *Book) AvailableByAsset(asset domain.Asset) []domain.Balance {
	out := make([]domain.Balance, 0)
	for _, balance := range book.balances {
		if balance.Asset == asset && balance.Available > 0 {
			out = append(out, balance)
		}
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].Account < out[j].Account
	})
	return out
}

func (book *Book) ValidateNonNegative() []ReconciliationIssue {
	issues := make([]ReconciliationIssue, 0)
	for _, balance := range book.balances {
		if balance.Available < 0 {
			issues = append(issues, ReconciliationIssue{
				Code:    "negative_available",
				Account: balance.Account,
				Asset:   balance.Asset,
				Actual:  balance.Available,
				Message: fmt.Sprintf("account %s has negative available %s", balance.Account, balance.Asset),
			})
		}
		if balance.Reserved < 0 {
			issues = append(issues, ReconciliationIssue{
				Code:    "negative_reserved",
				Account: balance.Account,
				Asset:   balance.Asset,
				Actual:  balance.Reserved,
				Message: fmt.Sprintf("account %s has negative reserved %s", balance.Account, balance.Asset),
			})
		}
	}
	return issues
}
