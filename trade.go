package bittrex

import (
	"time"
)

// Trade represents a trade that is planned or that already happened using the
// Bittrex API.
type Trade struct {
	BaseCurrency   string
	MarketCurrency string

	Type     TradeType
	Price    float64
	Quantity float64
	Time     time.Time
}

// Market returns the market.
func (t *Trade) Market() string {
	return t.BaseCurrency + "-" + t.MarketCurrency
}

// TradeType represents a type of trade: either buy or sell.
type TradeType int

const (
	// InvalidTradeType is used as the default for the Type type to indicate a
	// variable has not been explicitly set to one of the valid values.
	InvalidTradeType TradeType = iota

	// BuyType represents buy trades.
	BuyType

	// SellType represents sell trades.
	SellType
)
