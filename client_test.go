package bittrex_test

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/carterjones/bittrex"
)

func mustReadTestFixture(fixture string) []byte {
	data, err := ioutil.ReadFile(filepath.Join("test-fixtures", fixture))
	panicIfErr(err)
	return data
}

func panicIfErr(err error) {
	if err != nil {
		panic(err)
	}
}

var (
	fixtureAccountGetbalances     = mustReadTestFixture("account_getbalances.json")
	fixtureAccountGetorderhistory = mustReadTestFixture("account_getorderhistory.json")
	fixturePubGetticks            = mustReadTestFixture("pub_market_getticks.json")
	fixturePublicGetmarkets       = mustReadTestFixture("public_getmarkets.json")
)

func TestClient_Markets(t *testing.T) {
	cases := map[string]struct {
		client  *bittrex.Client
		exp     []bittrex.Market
		wantErr string
	}{
		"normal": {
			client: bittrex.New("", ""),
			exp: func() []bittrex.Market {
				var result struct {
					Result []bittrex.Market `json:"result"`
				}
				err := json.Unmarshal(fixturePublicGetmarkets, &result)
				panicIfErr(err)
				return result.Result
			}(),
		},
	}

	for id, tc := range cases {
		ts, _ := bittrex.NewMockRestServer()
		ts.Start()
		tc.client.HTTPClient = ts.Client()
		tc.client.HostAddr = ts.URL

		act, err := tc.client.Markets()
		if tc.wantErr != "" {
			errMatches(t, id, err, tc.wantErr)
		} else {
			equals(t, id, tc.exp, act)
			ok(t, id, err)
		}
	}
}

func TestClient_Balances(t *testing.T) {
	cases := map[string]struct {
		client  *bittrex.Client
		exp     []bittrex.Balance
		wantErr string
	}{
		"normal": {
			client: bittrex.New("", ""),
			exp: func() []bittrex.Balance {
				var result struct {
					Result []bittrex.Balance `json:"result"`
				}
				err := json.Unmarshal(fixtureAccountGetbalances, &result)
				panicIfErr(err)
				return result.Result
			}(),
		},
	}

	for id, tc := range cases {
		ts, _ := bittrex.NewMockRestServer()
		ts.Start()
		tc.client.HTTPClient = ts.Client()
		tc.client.HostAddr = ts.URL

		act, err := tc.client.Balances()
		if tc.wantErr != "" {
			errMatches(t, id, err, tc.wantErr)
		} else {
			equals(t, id, tc.exp, act)
			ok(t, id, err)
		}
	}
}

func TestClient_OrderHistory(t *testing.T) {
	cases := map[string]struct {
		client  *bittrex.Client
		exp     []bittrex.Order
		wantErr string
	}{
		"normal": {
			client: bittrex.New("", ""),
			exp: func() []bittrex.Order {
				var result struct {
					Result []bittrex.Order `json:"result"`
				}
				err := json.Unmarshal(fixtureAccountGetorderhistory, &result)
				panicIfErr(err)
				return result.Result
			}(),
		},
	}

	for id, tc := range cases {
		ts, _ := bittrex.NewMockRestServer()
		ts.Start()
		tc.client.HTTPClient = ts.Client()
		tc.client.HostAddr = ts.URL

		act, err := tc.client.OrderHistory()
		if tc.wantErr != "" {
			errMatches(t, id, err, tc.wantErr)
		} else {
			equals(t, id, tc.exp, act)
			ok(t, id, err)
		}
	}
}

func TestClient_LimitSell(t *testing.T) {
	cases := map[string]struct {
		client   *bittrex.Client
		market   string
		quantity float64
		rate     float64
		wantErr  string
	}{
		"normal": {
			client:   bittrex.New("", ""),
			market:   "BTC-ETH",
			quantity: 9001.0,
			rate:     1000.9,
		},
	}

	for id, tc := range cases {
		ts, rr := bittrex.NewMockRestServer()
		ts.Start()
		tc.client.HTTPClient = ts.Client()
		tc.client.HostAddr = ts.URL

		err := tc.client.LimitSell(tc.market, tc.quantity, tc.rate)
		if tc.wantErr != "" {
			errMatches(t, id, err, tc.wantErr)
		} else {
			expQuantity := fmt.Sprintf("%.8f", tc.quantity)
			expRate := fmt.Sprintf("%.8f", tc.rate)
			equals(t, id, tc.market, rr.Params.Get("market"))
			equals(t, id, expQuantity, rr.Params.Get("quantity"))
			equals(t, id, expRate, rr.Params.Get("rate"))
			ok(t, id, err)
		}
	}
}

func TestClient_LimitBuy(t *testing.T) {
	cases := map[string]struct {
		client   *bittrex.Client
		market   string
		quantity float64
		rate     float64
		wantErr  string
	}{
		"normal": {
			client:   bittrex.New("", ""),
			market:   "BTC-ETH",
			quantity: 9001.0,
			rate:     1000.9,
		},
	}

	for id, tc := range cases {
		ts, rr := bittrex.NewMockRestServer()
		ts.Start()
		tc.client.HTTPClient = ts.Client()
		tc.client.HostAddr = ts.URL

		err := tc.client.LimitBuy(tc.market, tc.quantity, tc.rate)
		if tc.wantErr != "" {
			errMatches(t, id, err, tc.wantErr)
		} else {
			expQuantity := fmt.Sprintf("%.8f", tc.quantity)
			expRate := fmt.Sprintf("%.8f", tc.rate)
			equals(t, id, tc.market, rr.Params.Get("market"))
			equals(t, id, expQuantity, rr.Params.Get("quantity"))
			equals(t, id, expRate, rr.Params.Get("rate"))
			ok(t, id, err)
		}
	}
}

func TestClient_Cancel(t *testing.T) {
	cases := map[string]struct {
		client    *bittrex.Client
		orderUUID string
		wantErr   string
	}{
		"normal": {
			client: bittrex.New("", ""),
		},
	}

	for id, tc := range cases {
		ts, rr := bittrex.NewMockRestServer()
		ts.Start()
		tc.client.HTTPClient = ts.Client()
		tc.client.HostAddr = ts.URL

		err := tc.client.Cancel(tc.orderUUID)
		if tc.wantErr != "" {
			errMatches(t, id, err, tc.wantErr)
		} else {
			equals(t, id, tc.orderUUID, rr.Params.Get("uuid"))
			ok(t, id, err)
		}
	}
}

func TestClient_Ticks(t *testing.T) {
	cases := map[string]struct {
		client  *bittrex.Client
		market  string
		exp     []bittrex.Tick
		wantErr string
	}{
		"normal": {
			client: bittrex.New("", ""),
			market: "BTC-WAVES",
			exp: func() []bittrex.Tick {
				var result struct {
					Result []bittrex.Tick `json:"result"`
				}
				err := json.Unmarshal(fixturePubGetticks, &result)
				panicIfErr(err)
				return result.Result
			}(),
		},
	}

	for id, tc := range cases {
		ts, _ := bittrex.NewMockRestServer()
		ts.Start()
		tc.client.HTTPClient = ts.Client()
		tc.client.HostAddr = ts.URL

		act, err := tc.client.Ticks(tc.market)
		if tc.wantErr != "" {
			errMatches(t, id, err, tc.wantErr)
		} else {
			equals(t, id, tc.exp, act)
			ok(t, id, err)
		}
	}
}
