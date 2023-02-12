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
	if s == "lovelace" {
		return ADA, nil
	}
	as := strings.Split(s, ".")
	if len(as) == 1 {
		return Asset{
			CurrencySymbol: as[0],
			TokenName:      "",
		}, nil
	} else if len(as) == 2 {
		return Asset{
			CurrencySymbol: as[0],
			TokenName:      as[1],
		}, nil
	} else {
		return Asset{}, errors.New("cannot parse asset from string, expect input to have format lovelace, $policyID or $policyID.$assetName")
	}
}

func (as Asset) String() string {
	if as == ADA {
		return "lovelace"
	} else if as.TokenName == "" {
		return as.CurrencySymbol
	} else {
		return as.CurrencySymbol + "." + as.TokenName
	}
}
func (a Asset) Cmp(b Asset) int {
	if a.CurrencySymbol == b.CurrencySymbol {
		return strings.Compare(a.TokenName, b.TokenName)
	}
	return strings.Compare(a.CurrencySymbol, b.CurrencySymbol)
}
