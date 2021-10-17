package cli

import (
	"fmt"
	"math/big"
	"strconv"
	"strings"

	"github.com/minswap/pab-go/ledger"
)

// Example:
// 10000 lovelace
// 10000 currencysymbol.tokenname
func parseFund(s string) (c ledger.Asset, amount *big.Int, err error) {
	fields := strings.Split(s, " ")
	amount = big.NewInt(0)
	if _, ok := amount.SetString(fields[0], 10); !ok {
		return c, nil, fmt.Errorf("fail to parse integer amount: %s", fields[0])
	}
	if fields[1] == "lovelace" {
		c = ledger.ADA
	} else {
		assetParts := strings.Split(fields[1], ".")
		if len(assetParts) == 1 {
			c = ledger.NewAsset(assetParts[0], "")
		} else {
			c = ledger.NewAsset(assetParts[0], assetParts[1])
		}
	}
	return c, amount, nil
}

// Example:
// TxOutDatumHashNone
// TxOutDatumHash ScriptDataInAlonzoEra "9e1199a988ba72ffd6e9c269cadb3b53b5f360ff99f112d9b2ee30c4d74ad88b"
func parseDatumHash(s string) *string {
	if s == "TxOutDatumHashNone" {
		return nil
	}
	hash := s[len(s)-65 : len(s)-1]
	return &hash
}

func parseFundsAndDatum(s string) (funds ledger.Value, datumHash *string, err error) {
	fields := strings.Split(s, " + ")
	funds = ledger.NewValue()
	// Skip last field because it's datum
	for _, field := range fields[:len(fields)-1] {
		asset, amount, err := parseFund(field)
		if err != nil {
			return nil, nil, fmt.Errorf("fail to parse fund: %w", err)
		}
		funds.Add(asset, amount)
	}
	datumHash = parseDatumHash(fields[len(fields)-1])
	return funds, datumHash, err
}

func parseUtxoOutput(addr string, output string) (utxos []ledger.Utxo, err error) {
	defer func() {
		if r := recover(); r != nil {
			utxos = nil
			err = fmt.Errorf("panic when parse utxo output: %v", r)
		}
	}()

	lines := strings.Split(output, "\n")
	// Skip first 2 header lines
	lines = lines[2:]
	for _, line := range lines {
		// Skip blank line
		if strings.TrimSpace(line) == "" {
			continue
		}
		fields := strings.Fields(line)
		txId := fields[0]
		txIndex, err := strconv.Atoi(fields[1])
		if err != nil {
			return nil, fmt.Errorf("fail to parse tx index: %w", err)
		}

		funds, datumHash, err := parseFundsAndDatum(strings.Join(fields[2:], " "))
		if err != nil {
			return nil, fmt.Errorf("fail to parse funds and datum hash: %w", err)
		}

		utxo := ledger.Utxo{
			TxID:      txId,
			TxIndex:   txIndex,
			Address:   addr,
			Value:     funds,
			DatumHash: datumHash,
		}
		utxos = append(utxos, utxo)
	}
	return utxos, nil
}
