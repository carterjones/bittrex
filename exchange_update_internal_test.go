package bittrex

import (
	"testing"
	"time"
)

func TestExchangeUpdate_Time(t *testing.T) {
	cases := map[string]struct {
		in      tradeFill
		exp     time.Time
		wantErr string
	}{
		"valid timestamp": {
			in:  tradeFill{TimeStamp: "2014-07-09T02:58:46.000"},
			exp: time.Date(2014, 07, 9, 2, 58, 46, 000, time.UTC),
		},
		"wrong format timestamp": {
			in:      tradeFill{TimeStamp: "2014-07-09 02:58:46.271"},
			wantErr: "time parse failed",
		},
		"not even a valid time identifier": {
			in:      tradeFill{TimeStamp: "faketimestamp"},
			wantErr: "time parse failed",
		},
	}

	for id, tc := range cases {
		act, err := tc.in.time()
		if tc.wantErr == "" {
			equals(t, id, tc.exp, act)
			ok(t, id, err)
		} else {
			errMatches(t, id, err, tc.wantErr)
		}
	}
}
