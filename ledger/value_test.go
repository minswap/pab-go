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
