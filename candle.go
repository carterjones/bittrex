package bittrex

import (
	"fmt"
	"strconv"
	"time"
)

// Candle contains the open, high, low, close, and volume data for a market.
type Candle struct {
	Market string
	Time   time.Time
	Open   float64
	High   float64
	Low    float64
	Close  float64
	Volume float64
}

func (candle Candle) String() string {
	o := strconv.FormatFloat(candle.Open, 'f', 8, 64)
	h := strconv.FormatFloat(candle.High, 'f', 8, 64)
	l := strconv.FormatFloat(candle.Low, 'f', 8, 64)
	c := strconv.FormatFloat(candle.Close, 'f', 8, 64)
	v := strconv.FormatFloat(candle.Volume, 'f', 8, 64)

	return fmt.Sprintf("%s: %s|O:%s|H:%s|L:%s|C:%s|V:%s", candle.Market, candle.Time.Format(time.RFC3339), o, h, l, c, v)
}
