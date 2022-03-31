package ledger

import (
	"errors"
	"strings"
)

var ADA = NewAsset("", "")

type Asset struct {
	CurrencySymbol string `json:"currencySymbol"`
	TokenName      string `json:"tokenName"`
}

func NewAsset(currencySymbol, tokenName string) Asset {
	return Asset{currencySymbol, tokenName}
}

func AssetFromString(s string) (Asset, error) {
	parts := strings.Split(s, ".")
	if len(parts) != 2 {
		return Asset{}, errors.New("ledger.AssetFromString: asset format must be $currencySymbol.$tokenName")
	}
	return NewAsset(parts[0], parts[1]), nil
}

func (a Asset) Cmp(b Asset) int {
	if a.CurrencySymbol == b.CurrencySymbol {
		return strings.Compare(a.TokenName, b.TokenName)
	}
	return strings.Compare(a.CurrencySymbol, b.CurrencySymbol)
}

func (a Asset) String() string {
	return a.CurrencySymbol + "." + a.TokenName
}
