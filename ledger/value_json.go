package ledger

import (
	"encoding/json"
	"math/big"
)

type ValueJSONItem struct {
	Asset  Asset    `json:"asset"`
	Amount *big.Int `json:"amount"`
}

type ValueJSON []ValueJSONItem

func (v Value) MarshalJSON() ([]byte, error) {
	if v == nil {
		return []byte("null"), nil
	}
	arr := make(ValueJSON, 0)
	for asset, amount := range v {
		arr = append(arr, ValueJSONItem{asset, amount})
	}
	return json.Marshal(arr)
}

func (v *Value) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		return nil
	}
	var arr ValueJSON
	if err := json.Unmarshal(data, &arr); err != nil {
		return err
	}
	*v = NewValue()
	for _, item := range arr {
		v.Add(item.Asset, item.Amount)
	}
	return nil
}
