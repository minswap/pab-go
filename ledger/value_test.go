package ledger

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValueClone(t *testing.T) {
	val1 := NewValue().Add(ADA, big.NewInt(1_000_000))
	val2 := val1.Clone().Remove(ADA, big.NewInt(300_000))

	assert.Equal(t, val1[ADA].Cmp(big.NewInt(1_000_000)), 0)
	assert.Equal(t, val2[ADA].Cmp(big.NewInt(700_000)), 0)
}

func TestMinimumADA(t *testing.T) {
	assert.Equal(t,
		NewValue().
			Add(ADA, big.NewInt(1250000000)).
			Add(
				NewAsset("1d7f33bd23d85e1a25d87d86fac4f199c3197a2f7afeb662a0f34e1e", "776f726c646d6f62696c65746f6b656e"),
				big.NewInt(1491513653)).
			Add(
				NewAsset("3f6092645942a54a75186b25e0975b79e1f50895ad958b42015eb6d2", "4d494e53574150"),
				big.NewInt(1),
			).
			Add(
				NewAsset("5178cc70a14405d3248e415d1a120c61d2aa74b4cee716d475b1495e", "3d6e0553e80f44a201b15eba1d31666083adc505e738efcccd84d464200183a7"),
				big.NewInt(1),
			).
			MinimumADA(true).
			Int64(),
		int64(2241330),
	)
}

func TestTrimValue(t *testing.T) {
	testAsset := NewAsset("1d7f33bd23d85e1a25d87d86fac4f199c3197a2f7afeb662a0f34e1e", "776f726c646d6f62696c65746f6b656e")
	originVal := NewValue().
		Add(ADA, big.NewInt(1000)).
		Add(testAsset, big.NewInt(0))
	trimmedVal := TrimValue(originVal)
	assert.Equal(t,
		trimmedVal[ADA].Int64(), int64(1000))
	assert.Nil(t,
		trimmedVal[testAsset])
}
