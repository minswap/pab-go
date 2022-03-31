package ledger

import (
	"errors"
	"fmt"
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
	if s == "lovelace" {
		return ADA, nil
	}
	as := strings.Split(s, ".")
	if len(as) != 2 {
		return Asset{}, errors.New("cannot parse asset from string")
	}
	return Asset{
		CurrencySymbol: as[0],
		TokenName:      as[1],
	}, nil
}

func (as Asset) String() string {
	if as == ADA {
		return "lovelace"
	} else {
		return fmt.Sprintf(`%s.%s`, as.CurrencySymbol, as.TokenName)
	}
}
func (a Asset) Cmp(b Asset) int {
	if a.CurrencySymbol == b.CurrencySymbol {
		return strings.Compare(a.TokenName, b.TokenName)
	}
	return strings.Compare(a.CurrencySymbol, b.CurrencySymbol)
}
