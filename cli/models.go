package cli

type cborFile struct {
	Type        string `json:"type"`
	Description string `json:"description"`
	CBORHex     string `json:"cborHex"`
}

type Tx struct {
	TxHash string `json:"txHash"`
	TxBody string `json:"txBody"`
}

type Tip struct {
	Epoch        int    `json:"epoch"`
	Hash         string `json:"hash"`
	Slot         int    `json:"slot"`
	Block        int    `json:"block"`
	Era          string `json:"era"`
	SyncProgress string `json:"syncProgress"`
}
