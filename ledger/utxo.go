package ledger

import "errors"

type Utxo struct {
	TxID      string  `json:"txID"`
	TxIndex   int     `json:"txIndex"`
	Address   string  `json:"address"`
	Value     Value   `json:"value"`
	DatumHash *string `json:"datumHash"`
}

var ErrNoCollateral = errors.New("no suitable collateral")

// FindCollateral find the biggest only-ADA UTxO
func FindCollateral(utxos []Utxo) (Utxo, error) {
	var ret Utxo
	found := false
	for _, u := range utxos {
		if len(u.Value) == 1 && u.Value.Contains(ADA) {
			if !found || u.Value[ADA].Cmp(ret.Value[ADA]) > 0 {
				ret = u
				found = true
			}
		}
	}
	if !found {
		return ret, ErrNoCollateral
	}
	return ret, nil
}

func SumValueOfUtxos(utxos []Utxo) Value {
	val := NewValue()
	for _, u := range utxos {
		val.AddAll(u.Value)
	}
	return val
}
