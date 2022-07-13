package cli

import (
	"fmt"
	"log"
	"math/big"
	"strconv"
	"strings"
	"unicode"

	"github.com/minswap/pab-go/ledger"
	"github.com/minswap/pab-go/txbuilder"
)

// Example: 5ca53b0eb10f317a5f1bf1bda679a04a8dd01c156643deee1d406ec2cc15e9e3#0
func buildInput(u txbuilder.TxInput) string {
	return fmt.Sprintf("%s#%d", u.TxID, u.TxIndex)
}

// Example: 100 lovelace + 42 5ca53b0eb10f317a5f1bf1bda679a04a8dd01c156643deee1d406ec2cc15e9e3.foo
func buildValue(val ledger.Value) string {
	var parts []string
	for c, amount := range val {
		if c == ledger.ADA {
			parts = append(parts, fmt.Sprintf("%s lovelace", amount.String()))
		} else if c.TokenName == "" {
			parts = append(parts, fmt.Sprintf("%s %s", amount.String(), c.CurrencySymbol))
		} else {
			parts = append(parts, fmt.Sprintf("%s %s.%s", amount.String(), c.CurrencySymbol, c.TokenName))
		}
	}
	return strings.Join(parts, " + ")
}

func buildExUnits(exCPU, exMem int64) string {
	return fmt.Sprintf("(%d,%d)", exCPU, exMem)
}

func buildOutput(o txbuilder.TxOutput) string {
	return fmt.Sprintf("%s + %s", o.Address, buildValue(o.Value))
}

func (cli *CardanoCLI) buildTempFile(suffix string, content string, temp *TempManager) string {
	file := temp.NewFile(suffix)
	file.WriteString(content)
	if cli.LogTempFile {
		// remove all whitespaces
		compactContent := strings.Map(func(r rune) rune {
			if !unicode.IsSpace(r) {
				return r
			}
			return -1
		}, content)
		log.Printf("%s has content: %s\n", file.Name(), compactContent)
	}
	return file.Name()
}

func (cli *CardanoCLI) buildTx(b txbuilder.TxBuilder, temp *TempManager) []string {
	var args []string
	eraFlag := ""
	switch cli.Era {
	case Alonzo:
		{
			eraFlag = "--alonzo-era"
			break
		}
	case Babbage:
		{
			eraFlag = "--babbage-era"
			break
		}
	}
	if b.IsRaw() {
		args = []string{"transaction", "build-raw", eraFlag, "--fee", strconv.FormatInt(b.Fee, 10)}
	} else {
		args = []string{"transaction", "build", eraFlag, "--change-address", b.ChangeAddress}
	}

	// build inputs
	for _, in := range b.PubKeyInputs {
		args = append(args, "--tx-in", buildInput(in))
	}

	for _, in := range b.ScriptInputs {
		args = append(args,
			"--tx-in", buildInput(in.TxInput),
			"--tx-in-script-file", in.ScriptFilePath,
			"--tx-in-datum-file", cli.buildTempFile("input-datum", in.DatumValue, temp),
			"--tx-in-redeemer-file", cli.buildTempFile("input-redeemer", in.RedeemerValue, temp),
		)
		if b.IsRaw() {
			args = append(args, "--tx-in-execution-units", buildExUnits(in.ExCPU, in.ExMem))
		}
	}
	for _, col := range b.Collaterals {
		args = append(args, "--tx-in-collateral", buildInput(col))
	}

	// build outputs
	for _, out := range b.PubKeyOutputs {
		args = append(args, "--tx-out", buildOutput(out))
	}
	for _, out := range b.ScriptOutputs {
		args = append(args,
			"--tx-out", buildOutput(out.TxOutput),
		)
		switch datum := out.Datum.(type) {
		case txbuilder.ScriptOutputDatumHash:
			args = append(args,
				"--tx-out-datum-hash", datum.DatumHash,
			)
		case txbuilder.ScriptOutputDatumValue:
			args = append(args,
				"--tx-out-datum-embed-file", cli.buildTempFile("output-datum-embed", datum.DatumValue, temp),
			)
		default:
			panic(fmt.Sprintf("Unsupported datum type: %T", datum))
		}
	}

	// build minting and burning
	forgeVal := ledger.NewValue()
	for _, mint := range b.Minting {
		forgeVal.AddAll(mint.Value)
		args = append(args,
			"--mint-script-file", mint.ScriptFilePath,
			"--mint-redeemer-file", cli.buildTempFile("mint-redeemer", mint.RedeemerValue, temp),
		)
		if b.IsRaw() {
			args = append(args, "--mint-execution-units", buildExUnits(mint.ExCPU, mint.ExMem))
		}
	}
	mintScriptFilePaths := make(map[string]struct{}, 0)
	for _, mintNativeScript := range b.MintingNativeScript {
		forgeVal.AddAll(mintNativeScript.Value)
		if _, ok := mintScriptFilePaths[mintNativeScript.ScriptFilePath]; !ok {
			mintScriptFilePaths[mintNativeScript.ScriptFilePath] = struct{}{}
		}
	}
	for _, burn := range b.Burning {
		for asset, amount := range burn.Value {
			forgeVal.Add(asset, new(big.Int).Neg(amount))
		}
		args = append(args,
			"--mint-script-file", burn.ScriptFilePath,
			"--mint-redeemer-file", cli.buildTempFile("mint-redeemer", burn.RedeemerValue, temp),
		)
		if b.IsRaw() {
			args = append(args, "--mint-execution-units", buildExUnits(burn.ExCPU, burn.ExMem))
		}
	}
	for _, burnNativeScript := range b.BurningNativeScript {
		for asset, amount := range burnNativeScript.Value {
			forgeVal.Add(asset, new(big.Int).Neg(amount))
		}
		if _, ok := mintScriptFilePaths[burnNativeScript.ScriptFilePath]; !ok {
			mintScriptFilePaths[burnNativeScript.ScriptFilePath] = struct{}{}
		}
	}

	for mintScriptFilePath := range mintScriptFilePaths {
		args = append(args,
			"--mint-script-file", mintScriptFilePath,
		)
	}

	if len(forgeVal) > 0 {
		args = append(args, "--mint", buildValue(forgeVal))
	}
	if b.ValidRangeFrom != nil {
		args = append(args,
			"--invalid-before", fmt.Sprintf("%d", *b.ValidRangeFrom),
		)
	}
	if b.ValidRangeTo != nil {
		args = append(args,
			"--invalid-hereafter", fmt.Sprintf("%d", *b.ValidRangeTo),
		)
	}

	if b.JSONMetadata != "" {
		args = append(args,
			"--metadata-json-file", cli.buildTempFile("metadata-json", b.JSONMetadata, temp),
		)
	}

	for _, skey := range b.SignerSkeyPaths {
		args = append(args,
			"--required-signer", skey,
		)
	}
	for _, vkh := range b.RequiredSignerVkeyHashes {
		args = append(args,
			"--required-signer-hash", vkh,
		)
	}

	args = append(args, "--protocol-params-file", cli.ProtocolParamsPath)
	return args
}
