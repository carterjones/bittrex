package bittrex_test

import (
	"testing"
	"time"

	"github.com/carterjones/bittrex"
)

func TestTick_String(t *testing.T) {
	cases := map[string]struct {
		in  bittrex.Tick
		exp string
	}{
		"normal": {
			in: bittrex.Tick{
				Timestamp: "faketimestamp",
				Open:      2.1,
				High:      3.0,
				Low:       1.0,
				Close:     2.2,
				Volume:    9000.1,
			},
			exp: "faketimestamp: O:2.10000000, H:3.00000000, L:1.00000000, C:2.20000000, V:9000.10000000",
		},
	}

	for id, tc := range cases {
		act := tc.in.String()
		equals(t, id, tc.exp, act)
	}
}

func TestTick_Time(t *testing.T) {
	cases := map[string]struct {
		in      bittrex.Tick
		exp     time.Time
		wantErr string
	}{
		"valid timestamp": {
			in:  bittrex.Tick{Timestamp: "2017-12-22T03:15:49"},
			exp: time.Date(2017, 12, 22, 3, 15, 49, 0, time.UTC),
		},
		"wrong format timestamp": {
			in:      bittrex.Tick{Timestamp: "2017-12-22 03:15:49"},
			wantErr: "time parse failed",
		},
		"not even a valid time identifier": {
			in:      bittrex.Tick{Timestamp: "faketimestamp"},
			wantErr: "time parse failed",
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
