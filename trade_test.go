package bittrex_test

import (
	"testing"

	"github.com/carterjones/bittrex"
)

func TestTrade_Market(t *testing.T) {
	cases := map[string]struct {
		in  bittrex.Trade
		exp string
	}{
		"normal": {
			in:  bittrex.Trade{BaseCurrency: "BTC", MarketCurrency: "LTC"},
			exp: "BTC-LTC",
		},
		"lowercase": {
			in:  bittrex.Trade{BaseCurrency: "btc", MarketCurrency: "ltc"},
			exp: "btc-ltc",
		},
		"mixed case": {
			in:  bittrex.Trade{BaseCurrency: "bTc", MarketCurrency: "LtC"},
			exp: "bTc-LtC",
		},
	}

	for id, tc := range cases {
		act := tc.in.Market()
		equals(t, id, tc.exp, act)
	}
}
