package txbuilder

import (
	"github.com/minswap/pab-go/ledger"
)

type TxInput struct {
	TxID    string
	TxIndex int
	TxOut   TxOutput
}

type ScriptInput struct {
	TxInput
	ScriptFilePath string
	DatumValue     string
	RedeemerValue  string
	TxOut          ScriptOutput
	ExMem          int64
	ExCPU          int64
}

type TxOutput struct {
	Address string
	Value   ledger.Value
}

type ScriptOutput struct {
	TxOutput
	DatumHash string
}

type Minting struct {
	Value          ledger.Value
	ScriptFilePath string
	RedeemerValue  string
	ExMem          int64
	ExCPU          int64
}

type Burning struct {
	Value          ledger.Value
	ScriptFilePath string
	RedeemerValue  string
	ExMem          int64
	ExCPU          int64
}

type TxBuilder struct {
	PubKeyInputs   []TxInput
	ScriptInputs   []ScriptInput
	PubKeyOutputs  []TxOutput
	ScriptOutputs  []ScriptOutput
	Minting        []Minting
	Burning        []Burning
	ChangeAddress  string
	Fee            int64
	Collaterals    []TxInput
	ValidRangeFrom *int64
	ValidRangeTo   *int64
}

type Option = func(b *TxBuilder)

func New(opts ...Option) TxBuilder {
	b := TxBuilder{}
	for _, opt := range opts {
		opt(&b)
	}
	return b
}

func (b *TxBuilder) Add(opts ...Option) {
	for _, opt := range opts {
		opt(b)
	}
}

func (b *TxBuilder) IsRaw() bool {
	return b.Fee > 0
}

func SpendPubKeyUtxos(utxos ...ledger.Utxo) Option {
	return func(b *TxBuilder) {
		for _, u := range utxos {
			b.PubKeyInputs = append(b.PubKeyInputs, TxInput{
				TxID:    u.TxID,
				TxIndex: u.TxIndex,
				TxOut: TxOutput{
					Address: u.Address,
					Value:   u.Value,
				},
			})
		}
	}
}

func SpendScriptUtxo(u ledger.Utxo, scriptFilePath, datum, redeemer string) Option {
	return func(b *TxBuilder) {
		b.ScriptInputs = append(b.ScriptInputs, ScriptInput{
			TxInput: TxInput{
				TxID:    u.TxID,
				TxIndex: u.TxIndex,
			},
			ScriptFilePath: scriptFilePath,
			DatumValue:     datum,
			RedeemerValue:  redeemer,
			TxOut: ScriptOutput{
				TxOutput: TxOutput{
					Address: u.Address,
					Value:   u.Value,
				},
				DatumHash: *u.DatumHash,
			},
		})
	}
}

func SpendScriptUtxoRaw(u ledger.Utxo, scriptFilePath, datum, redeemer string, exMem, exCPU int64) Option {
	return func(b *TxBuilder) {
		b.ScriptInputs = append(b.ScriptInputs, ScriptInput{
			TxInput: TxInput{
				TxID:    u.TxID,
				TxIndex: u.TxIndex,
			},
			ScriptFilePath: scriptFilePath,
			DatumValue:     datum,
			RedeemerValue:  redeemer,
			ExMem:          exMem,
			ExCPU:          exCPU,
		})
	}
}

func MintAssets(val ledger.Value, scriptFilePath, redeemer string) Option {
	return func(b *TxBuilder) {
		b.Minting = append(b.Minting, Minting{
			Value:          val,
			ScriptFilePath: scriptFilePath,
			RedeemerValue:  redeemer,
		})
	}
}

func MintAssetsRaw(val ledger.Value, scriptFilePath, redeemer string, exMem, exCPU int64) Option {
	return func(b *TxBuilder) {
		b.Minting = append(b.Minting, Minting{
			Value:          val,
			ScriptFilePath: scriptFilePath,
			RedeemerValue:  redeemer,
			ExMem:          exMem,
			ExCPU:          exCPU,
		})
	}
}

func BurnAssets(val ledger.Value, scriptFilePath, redeemer string) Option {
	return func(b *TxBuilder) {
		b.Burning = append(b.Burning, Burning{
			Value:          val,
			ScriptFilePath: scriptFilePath,
			RedeemerValue:  redeemer,
		})
	}
}

func BurnAssetsRaw(val ledger.Value, scriptFilePath, redeemer string, exMem, exCPU int64) Option {
	return func(b *TxBuilder) {
		b.Burning = append(b.Burning, Burning{
			Value:          val,
			ScriptFilePath: scriptFilePath,
			RedeemerValue:  redeemer,
			ExMem:          exMem,
			ExCPU:          exCPU,
		})
	}
}

func PayToPubKey(addr string, val ledger.Value) Option {
	return func(b *TxBuilder) {
		b.PubKeyOutputs = append(b.PubKeyOutputs, TxOutput{
			Address: addr,
			Value:   val,
		})
	}
}

func PayToScript(addr string, val ledger.Value, datumHash string) Option {
	return func(b *TxBuilder) {
		b.ScriptOutputs = append(b.ScriptOutputs, ScriptOutput{
			TxOutput: TxOutput{
				Address: addr,
				Value:   val,
			},
			DatumHash: datumHash,
		})
	}
}

func PayChangeTo(addr string) Option {
	return func(b *TxBuilder) {
		b.ChangeAddress = addr
	}
}

func UseCollaterals(utxos ...ledger.Utxo) Option {
	return func(b *TxBuilder) {
		for _, u := range utxos {
			b.Collaterals = append(b.Collaterals, TxInput{
				TxID:    u.TxID,
				TxIndex: u.TxIndex,
			})
		}
	}
}

func PayFee(x int64) Option {
	return func(b *TxBuilder) {
		b.Fee = x
	}
}

func SetValidRangeFrom(from int64) Option {
	return func(b *TxBuilder) {
		b.ValidRangeFrom = &from
	}
}

func SetValidRangeTo(to int64) Option {
	return func(b *TxBuilder) {
		b.ValidRangeTo = &to
	}
}
