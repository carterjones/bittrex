package bittrex

import (
	"time"

	"github.com/pkg/errors"
)

type exchangeUpdate struct {
	MarketName string
	Nounce     uint // Note: this may be a typo for "Nonce"
	Buys       []tradeOrder
	Sells      []tradeOrder
	Fills      []tradeFill
}

type tradeOrder struct {
	Type     int
	Rate     float64
	Quantity float64
}

type tradeFill struct {
	OrderType string
	Rate      float64
	Quantity  float64
	TimeStamp string
}

func (tf *tradeFill) time() (time.Time, error) {
	layout := "2006-01-02T15:04:05.999"
	t, err := time.Parse(layout, tf.TimeStamp)
	if err != nil {
		return time.Time{}, errors.Wrap(err, "time parse failed")
	}
	return t, nil
}
