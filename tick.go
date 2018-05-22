package bittrex

import (
	"strconv"
	"time"

	"github.com/pkg/errors"
)

// Tick represents the open, high, low, and close values along with some other
// bits of data received from Bittrex.
//
// For example, many of these ticks will be returned by navigating to the
// following website:
// https://bittrex.com/Api/v2.0/pub/market/GetTicks?marketName=BTC-WAVES&tickInterval=oneMin
//
// Here is an example of what they should look like:
//
//  {
//    "O":0.00061830,
//    "H":0.00061830,
//    "L":0.00061798,
//    "C":0.00061798,
//    "V":1220.69744635,
//    "T":"2017-11-17T16:51:00",
//    "BV":0.75448216
//  }
type Tick struct {
	Open      float64 `json:"O"`
	High      float64 `json:"H"`
	Low       float64 `json:"L"`
	Close     float64 `json:"C"`
	Volume    float64 `json:"V"`
	Timestamp string  `json:"T"`
	BV        float64
}

func (t Tick) String() string {
	return t.Timestamp + ": " +
		"O:" + strconv.FormatFloat(t.Open, 'f', 8, 64) +
		", H:" + strconv.FormatFloat(t.High, 'f', 8, 64) +
		", L:" + strconv.FormatFloat(t.Low, 'f', 8, 64) +
		", C:" + strconv.FormatFloat(t.Close, 'f', 8, 64) +
		", V:" + strconv.FormatFloat(t.Volume, 'f', 8, 64)
}

// Time returns a time object that is converted from the timestamp value.
func (t *Tick) Time() (time.Time, error) {
	layout := "2006-01-02T15:04:05"
	tickTime, err := time.Parse(layout, t.Timestamp)
	if err != nil {
		return time.Time{}, errors.Wrap(err, "time parse failed")
	}
	return tickTime, nil
}
