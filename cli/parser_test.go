package cli

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseQueryUtxoOutput(t *testing.T) {
	testcase := `{
		"52db772a86ccc918a71ed3a6881692010f05f29b2839e7dc4c5ca95c129261fa#0": {
			"address": "addr_test1qzq29qg6d4m5w52e4w06ejn6l85ekthvnudcr6y5c7ka0q404j5fcr8xjh6djzmhkjuy2erva0f8dtvuz247tg2tz73snk0rtt",
			"value": {
				"c13eaa5804a65587ec36db51d21bcd8847efea3627e8a07e12cf304b": {
					"tMIN": 5000000000000000,
					"": 45000000000000000
				},
				"lovelace": 3000000000000
			}
		},
		"e18888b1f8559e59f479e72ee3f7e02dca395f5ee6f6b9a84ed67ffc02473d5d#2": {
			"address": "addr_test1wr37myp6qxqjd5g2de002z27zecggjfwqgdwn0wav8m4y3ggavlh3",
			"data": "7bb1486cc7fdd7b16f42afcec19183c5f2bb5f92c43cc3e2c9c56de5bb390116",
			"value": {
					"lovelace": 2007000000
			}
		}
	}`
	_, err := parseQueryUtxoOutput([]byte(testcase))
	assert.NoError(t, err)
}
