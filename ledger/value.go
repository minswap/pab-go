package ledger

import (
	"math/big"
)

const (
	utxoEntrySizeWithoutVal = 27
	utxoCostPerWord         = 34482
	justDataHashSize        = 10
	nothingSize             = 0
	policyLength            = 28
)

// Value is a list of assets and their amounts.
type Value map[Asset]*big.Int

func NewValue() Value {
	return make(Value)
}

func (v Value) Clone() Value {
	w := NewValue()
	w.AddAll(v)
	return w
}

func (v Value) Add(asset Asset, amount *big.Int) {
	if _, ok := v[asset]; !ok {
		v[asset] = big.NewInt(0)
	}
	v[asset].Add(v[asset], amount)
}

func (v Value) AddMinimumADA(isScriptUtxo bool) {
	min := v.MinimumADA(isScriptUtxo)
	if !v.Contains(ADA) || v[ADA].Cmp(min) < 0 {
		v[ADA] = min
	}
}

// Remove subtract asset amount in Value and remove asset if amount is negative
func (v Value) Remove(asset Asset, amount *big.Int) {
	if _, ok := v[asset]; !ok {
		return
	}
	v[asset].Sub(v[asset], amount)
	if v[asset].Cmp(big.NewInt(0)) <= 0 {
		v.RemoveAsset(asset)
	}
}

func (v Value) AddAll(w Value) {
	for asset, amount := range w {
		v.Add(asset, amount)
	}
}

func (v Value) RemoveAll(w Value) {
	for asset, amount := range w {
		v.Remove(asset, amount)
	}
}

func (v Value) RemoveAsset(c Asset) {
	delete(v, c)
}

func (v Value) Contains(c Asset) bool {
	_, ok := v[c]
	return ok
}

func (val Value) MinimumADA(isScriptUtxo bool) *big.Int {
	newVal := val.Clone()
	newVal.RemoveAsset(ADA)
	var policyIds = make(map[string]struct{})
	var tokenNames = make(map[string]struct{})

	for asset := range newVal {
		policyIds[asset.CurrencySymbol] = struct{}{}
		tokenNames[asset.TokenName] = struct{}{}
	}
	assetsSize := len(newVal) * 12
	sumAssetNameLength := 0
	for tn := range tokenNames {
		sumAssetNameLength += len(tn)
	}

	policyLength := len(policyIds) * policyLength
	valueSize := 6 + (assetsSize+sumAssetNameLength+policyLength+7)/8
	datumHashSize := 0
	if isScriptUtxo {
		datumHashSize = justDataHashSize
	} else {
		datumHashSize = nothingSize
	}

	utxoEntrySize := utxoEntrySizeWithoutVal + valueSize + datumHashSize
	minLovelace := utxoEntrySize * utxoCostPerWord

	return big.NewInt(int64(minLovelace))
}
