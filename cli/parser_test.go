package cli

import (
	"fmt"
	"math/big"
	"testing"

	"github.com/minswap/pab-go/ledger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseFund(t *testing.T) {
	t.Parallel()
	tests := []struct {
		in        string
		outAsset  ledger.Asset
		outAmount *big.Int
	}{
		{
			"10000 lovelace",
			ledger.ADA,
			big.NewInt(10000),
		},
		{
			"10000 9e1199a988ba72ffd6e9c269cadb3b53b5f360ff99f112d9b2ee30c4d74ad88b.MIN",
			ledger.NewAsset("9e1199a988ba72ffd6e9c269cadb3b53b5f360ff99f112d9b2ee30c4d74ad88b", "MIN"),
			big.NewInt(10000),
		},
		{
			"10000 9e1199a988ba72ffd6e9c269cadb3b53b5f360ff99f112d9b2ee30c4d74ad88b",
			ledger.NewAsset("9e1199a988ba72ffd6e9c269cadb3b53b5f360ff99f112d9b2ee30c4d74ad88b", ""),
			big.NewInt(10000),
		},
	}
	for i, tt := range tests {
		tt := tt
		t.Run(fmt.Sprintf("TestParseFund#%v", i), func(t *testing.T) {
			t.Parallel()
			asset, amount, err := parseFund(tt.in)
			require.NoError(t, err)
			assert.Equal(t, tt.outAsset, asset)
			assert.Equal(t, tt.outAmount.Cmp(amount), 0)
		})
	}
}

func TestParseDatumHash(t *testing.T) {
	assert.Nil(t, nil, parseDatumHash("TxOutDatumHashNone"))
	assert.Equal(t,
		"9e1199a988ba72ffd6e9c269cadb3b53b5f360ff99f112d9b2ee30c4d74ad88b",
		*parseDatumHash("TxOutDatumHash ScriptDataInAlonzoEra \"9e1199a988ba72ffd6e9c269cadb3b53b5f360ff99f112d9b2ee30c4d74ad88b\""),
	)
}
