package bittrex

import (
	"fmt"
	"strconv"
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

func (t Trade) String() string {
	mkt := t.Market()
	typ := t.Type.String()
	tim := t.Time.String()
	prc := strconv.FormatFloat(t.Price, 'f', 8, 64)
	qty := strconv.FormatFloat(t.Quantity, 'f', 8, 64)
	return fmt.Sprintf("%s: %s | %s | quantity: %s | price: %s",
		mkt, typ, tim, prc, qty)
}

// TradeType represents a type of trade: either buy or sell.
type TradeType int

func (t TradeType) String() string {
	switch t {
	case BuyType:
		return "BUY"
	case SellType:
		return "SELL"
	default:
		return "<invalid trade type>"
	}
}

const (
	// InvalidTradeType is used as the default for the Type type to indicate a
	// variable has not been explicitly set to one of the valid values.
	InvalidTradeType TradeType = iota

	// BuyType represents buy trades.
	BuyType

	// SellType represents sell trades.
	SellType
)
