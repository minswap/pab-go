package ledger

import "strings"

var ADA = NewAsset("", "")

type Asset struct {
	CurrencySymbol string `json:"currencySymbol"`
	TokenName      string `json:"tokenName"`
}

func NewAsset(currencySymbol, tokenName string) Asset {
	return Asset{currencySymbol, tokenName}
}

func (a Asset) Cmp(b Asset) int {
	if a.CurrencySymbol == b.CurrencySymbol {
		return strings.Compare(a.TokenName, b.TokenName)
	}
	return strings.Compare(a.CurrencySymbol, b.CurrencySymbol)
}
