package ledger

type Tip struct {
	Epoch        int    `json:"epoch"`
	Hash         string `json:"hash"`
	Slot         int    `json:"slot"`
	Block        int    `json:"block"`
	Era          string `json:"era"`
	SyncProgress string `json:"syncProgress"`
}
