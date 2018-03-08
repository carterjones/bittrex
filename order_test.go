package bittrex_test

import (
	"testing"
	"time"

	"github.com/carterjones/bittrex"
)

func TestOrder_Market(t *testing.T) {
	cases := map[string]struct {
		in  bittrex.Order
		exp string
	}{
		"correct": {in: bittrex.Order{Exchange: "abc123"}, exp: "abc123"},
	}

	for id, tc := range cases {
		act := tc.in.Market()
		equals(t, id, tc.exp, act)
	}
}

func TestOrder_Price(t *testing.T) {
	cases := map[string]struct {
		in  bittrex.Order
		exp float64
	}{
		"correct": {in: bittrex.Order{PricePerUnit: 123.456}, exp: 123.456},
	}

	for id, tc := range cases {
		act := tc.in.Rate()
		equals(t, id, tc.exp, act)
	}
}

func TestOrder_Time(t *testing.T) {
	cases := map[string]struct {
		in      bittrex.Order
		exp     time.Time
		wantErr string
	}{
		"valid timestamp": {
			in:  bittrex.Order{TimeStamp: "2014-07-09T02:58:46.00"},
			exp: time.Date(2014, 07, 9, 2, 58, 46, 00, time.UTC),
		},
		"wrong format timestamp": {
			in:      bittrex.Order{TimeStamp: "2014-07-09 02:58:46.27"},
			wantErr: "failed to parse time",
		},
		"not even a valid time identifier": {
			in:      bittrex.Order{TimeStamp: "faketimestamp"},
			wantErr: "failed to parse time",
		},
	}

	for id, tc := range cases {
		act, err := tc.in.Time()
		if tc.wantErr == "" {
			equals(t, id, tc.exp, act)
			ok(t, id, err)
		} else {
			errMatches(t, id, err, tc.wantErr)
		}
	}
}
