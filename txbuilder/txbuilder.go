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

type ScriptOutputDatum interface {
	isScriptOutputDatum()
}

type ScriptOutputDatumHash struct {
	DatumHash string
}

func (ScriptOutputDatumHash) isScriptOutputDatum() {}

type ScriptOutputDatumValue struct {
	DatumValue string
}

func (ScriptOutputDatumValue) isScriptOutputDatum() {}

type ScriptOutput struct {
	TxOutput
	Datum ScriptOutputDatum
}

type Minting struct {
	Value          ledger.Value
	ScriptFilePath string
	RedeemerValue  string
	ExMem          int64
	ExCPU          int64
}

type MintingNativeScript struct {
	Value          ledger.Value
	ScriptFilePath string
}

type Burning struct {
	Value          ledger.Value
	ScriptFilePath string
	RedeemerValue  string
	ExMem          int64
	ExCPU          int64
}

type BurningNativeScript struct {
	Value          ledger.Value
	ScriptFilePath string
}

type TxBuilder struct {
	PubKeyInputs             []TxInput
	ScriptInputs             []ScriptInput
	PubKeyOutputs            []TxOutput
	ScriptOutputs            []ScriptOutput
	Minting                  []Minting
	MintingNativeScript      []MintingNativeScript
	Burning                  []Burning
	BurningNativeScript      []BurningNativeScript
	ChangeAddress            string
	Fee                      int64
	Collaterals              []TxInput
	ValidRangeFrom           *int64
	ValidRangeTo             *int64
	JSONMetadata             string
	SignerSkeyPaths          []string // TODO: Rename to RequiredSignerSkeyPaths in next breaking change
	RequiredSignerVkeyHashes []string
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
				Datum: ScriptOutputDatumHash{*u.DatumHash},
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

func MintNativeScriptAssets(val ledger.Value, scriptFilePath string) Option {
	return func(b *TxBuilder) {
		b.MintingNativeScript = append(b.MintingNativeScript, MintingNativeScript{
			Value:          val,
			ScriptFilePath: scriptFilePath,
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

func BurnNativeScriptAssets(val ledger.Value, scriptFilePath string) Option {
	return func(b *TxBuilder) {
		b.BurningNativeScript = append(b.BurningNativeScript, BurningNativeScript{
			Value:          val,
			ScriptFilePath: scriptFilePath,
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

func PayToScript(addr string, val ledger.Value, datum ScriptOutputDatum) Option {
	return func(b *TxBuilder) {
		b.ScriptOutputs = append(b.ScriptOutputs, ScriptOutput{
			TxOutput: TxOutput{
				Address: addr,
				Value:   val,
			},
			Datum: datum,
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

func SetJSONMetadata(json string) Option {
	return func(b *TxBuilder) {
		b.JSONMetadata = json
	}
}

// TODO: Rename to RequireSignWithSkey in next breaking change
func SignedWith(sKeyPaths ...string) Option {
	return func(b *TxBuilder) {
		b.SignerSkeyPaths = append(b.SignerSkeyPaths, sKeyPaths...)
	}
}

func RequireSignWithVkeyHash(vkeyHashes ...string) Option {
	return func(b *TxBuilder) {
		b.RequiredSignerVkeyHashes = append(b.RequiredSignerVkeyHashes, vkeyHashes...)
	}
}
