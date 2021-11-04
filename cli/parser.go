package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"os"
	"strconv"
	"strings"

	"github.com/minswap/pab-go/ledger"
)

type QueryUtxoOutFile = map[string]struct {
	Address string                     `json:"address"`
	Value   map[string]json.RawMessage `json:"value"`
	Data    *string                    `json:"data"`
}

func parseTxIdTxIx(input string) (txId string, txIx int, err error) {
	fields := strings.Split(input, "#")
	if len(fields) != 2 {
		return "", 0, fmt.Errorf("expect format txId#txIx, got %s", input)
	}
	txIx, err = strconv.Atoi(fields[1])
	if err != nil {
		return "", 0, fmt.Errorf("error when parse txIx: %s: %w", fields[1], err)
	}
	return fields[0], txIx, nil
}

func parseValue(input map[string]json.RawMessage) (ledger.Value, error) {
	val := ledger.NewValue()
	for currencySymbol, raw := range input {
		if currencySymbol == "lovelace" {
			amount, ok := new(big.Int).SetString(string(raw), 10)
			if !ok {
				return nil, fmt.Errorf("fail to parse lovelace amount: %s", raw)
			}
			val[ledger.ADA] = amount
			continue
		}

		var tokenNameMap map[string]*big.Int
		if err := json.Unmarshal(raw, &tokenNameMap); err != nil {
			return nil, fmt.Errorf("fail to decode map of token name to amount: %w", err)
		}
		for tokenName, amount := range tokenNameMap {
			val.Add(ledger.NewAsset(currencySymbol, tokenName), amount)
		}
	}
	return val, nil
}

func parseQueryUtxoOutput(outputBytes []byte) ([]ledger.Utxo, error) {
	var output QueryUtxoOutFile
	if err := json.Unmarshal(outputBytes, &output); err != nil {
		return nil, fmt.Errorf("fail to decode query utxo output: %w", err)
	}
	utxos := make([]ledger.Utxo, 0)
	for txIdTxIx, txOut := range output {
		txId, txIx, err := parseTxIdTxIx(txIdTxIx)
		if err != nil {
			return nil, fmt.Errorf("fail to parse txIdTxIx: %w", err)
		}
		val, err := parseValue(txOut.Value)
		if err != nil {
			return nil, fmt.Errorf("fail to parse txOut value: %w", err)
		}
		utxo := ledger.Utxo{
			TxID:      txId,
			TxIndex:   txIx,
			Address:   txOut.Address,
			Value:     val,
			DatumHash: txOut.Data,
		}
		utxos = append(utxos, utxo)
	}
	return utxos, nil
}

func parseQueryUtxoOutFile(file *os.File) ([]ledger.Utxo, error) {
	outputBytes, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("fail to read query utxo out file: %w", err)
	}
	return parseQueryUtxoOutput(outputBytes)
}
