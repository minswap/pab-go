package ledger

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAssetFromString(t *testing.T) {
	assert := assert.New(t)
	a, err := AssetFromString("lovelace")
	if assert.NoError(err) {
		assert.Equal(ADA, a)
	}
	a, err = AssetFromString("29d222ce763455e3d7a09a665ce554f00ac89d2e99a1a83d267170c6.4d494e")
	if assert.NoError(err) {
		assert.Equal(Asset{
			CurrencySymbol: "29d222ce763455e3d7a09a665ce554f00ac89d2e99a1a83d267170c6",
			TokenName:      "4d494e",
		}, a)
	}
	a, err = AssetFromString("29d222ce763455e3d7a09a665ce554f00ac89d2e99a1a83d267170c6")
	if assert.NoError(err) {
		assert.Equal(Asset{
			CurrencySymbol: "29d222ce763455e3d7a09a665ce554f00ac89d2e99a1a83d267170c6",
			TokenName:      "",
		}, a)
	}
}
