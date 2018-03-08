package bittrex

import (
	"encoding/json"
	"time"

	"github.com/pkg/errors"
)

// Order represents an order that is placed on Bittrex.
type Order struct { // nolint: maligned
	OrderUUID         string `json:"OrderUuid"`
	Exchange          string
	TimeStamp         string // format: 2014-07-09T02:58:46.27
	OrderType         string
	Limit             float64
	Quantity          float64
	QuantityRemaining float64
	Commission        float64
	Price             float64
	PricePerUnit      float64
	IsConditional     bool
	Condition         string
	ConditionTarget   *json.RawMessage
	ImmediateOrCancel bool
	Closed            string // This is a timestamp.
}

// Market returns this order's exchange. It is purely a convenience function.
func (o *Order) Market() string {
	return o.Exchange
}

// Rate returns this order's price per unit. It is purely a convenience
// function.
func (o *Order) Rate() float64 {
	return o.PricePerUnit
}

// Time returns this order's timestamp converted into a time.Time object.
func (o *Order) Time() (time.Time, error) {
	layout := "2006-01-02T15:04:05.99"
	t, err := time.Parse(layout, o.TimeStamp)
	if err != nil {
		return time.Time{}, errors.Wrap(err, "failed to parse time")
	}
	return t, nil
}
