package cli

import (
	"fmt"
	"log"
	"math/big"
	"strconv"
	"strings"

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
	if cli.Debug {
		log.Printf("%s has content: %s\n", file.Name(), content)
	}
	return file.Name()
}

func (cli *CardanoCLI) buildTx(b txbuilder.TxBuilder, temp *TempManager) []string {
	var args []string
	if b.IsRaw() {
		args = []string{"transaction", "build-raw", "--alonzo-era", "--fee", strconv.FormatInt(b.Fee, 10)}
	} else {
		args = []string{"transaction", "build", "--alonzo-era", "--change-address", b.ChangeAddress}
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
		if out.DatumValue != nil {
			args = append(args,
				"--tx-out-datum-embed-file", cli.buildTempFile("output-datum-embed", *out.DatumValue, temp),
			)
		} else {
			args = append(args,
				"--tx-out-datum-hash", out.DatumHash,
			)
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
	for _, skey := range b.SignerSkeyPaths {
		args = append(args,
			"--required-signer", skey,
		)
	}

	args = append(args, "--protocol-params-file", cli.ProtocolParamsPath)
	return args
}
