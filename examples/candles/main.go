package main

import (
	"fmt"
	"log"
	"strconv"
	"sync"
	"time"

	"github.com/carterjones/bittrex"
)

func main() {
	c := bittrex.New("", "")

	// Register a handler to funnel each trade to a trades channel.
	trades := make(chan bittrex.Trade)
	c.Register(func(t bittrex.Trade) { trades <- t })

	// Create a candle channel that collects candles for each minute of trade
	// data.
	candles := generateCandles(trades)

	// Subscribe to all Bittrex markets.
	subscribeToAllMarkets(c)

	// Print the candles.
	for c := range candles {
		log.Println(c)
	}
}

func subscribeToAllMarkets(c *bittrex.Client) {
	ms, err := c.Markets()
	panicIfErr(err)

	for _, m := range ms {
		err = c.Subscribe(m.Name, panicIfErr)
		panicIfErr(err)
	}
}

func panicIfErr(err error) {
	if err != nil {
		log.Panic(err)
	}
}

// OHLC contains the open, high, low, and close data for a market.
type OHLC struct {
	Market string
	Open   float64
	High   float64
	Low    float64
	Close  float64
}

func (ohlc OHLC) String() string {
	o := strconv.FormatFloat(ohlc.Open, 'f', 8, 64)
	h := strconv.FormatFloat(ohlc.High, 'f', 8, 64)
	l := strconv.FormatFloat(ohlc.Low, 'f', 8, 64)
	c := strconv.FormatFloat(ohlc.Close, 'f', 8, 64)

	return fmt.Sprintf("%s: O:%s|H:%s|L:%s|C:%s", ohlc.Market, o, h, l, c)
}

func generateCandles(trades <-chan bittrex.Trade) <-chan OHLC {
	ohlcs := make(map[string]OHLC)
	ohlcsMux := sync.Mutex{}
	candles := make(chan OHLC)

	// Start a goroutine that updates OHLC values as each trade comes in.
	go func() {
		for {
			t := <-trades

			ohlcsMux.Lock()

			var ohlc OHLC
			var ok bool
			if ohlc, ok = ohlcs[t.Market()]; !ok {
				// If the OHLC does not already exist, then create one.
				ohlcs[t.Market()] = OHLC{
					Market: t.Market(),
					Open:   t.Price,
					High:   t.Price,
					Low:    t.Price,
					Close:  t.Price,
				}

				// Move to the next trade.
				ohlcsMux.Unlock()
				continue
			}

			// If the OHLC does exist, then modify it as necessary.

			// Set the open value if none is set.
			if ohlc.Open == 0.0 {
				ohlc.Open = t.Price
			}

			// Set the high value.
			if t.Price > ohlc.High {
				ohlc.High = t.Price
			}

			// Set the low value.
			if t.Price < ohlc.Low {
				ohlc.Low = t.Price
			}

			// Set the last price. This will eventually be used as the close
			// value.
			ohlc.Close = t.Price

			// Save the new value.
			ohlcs[t.Market()] = ohlc

			ohlcsMux.Unlock()
		}
	}()

	// Start a goroutine that waits for a minute, prints the current OHLC
	// values, and then resets the OHLC map.
	go func() {
		for {
			// Wait for a minute.
			<-time.After(1 * time.Minute)

			// Lock the map.
			ohlcsMux.Lock()

			// Save the OHLC to the candles channel.
			for _, v := range ohlcs {
				candles <- v
			}

			// Reset the map.
			ohlcs = make(map[string]OHLC)

			// Unlock the map.
			ohlcsMux.Unlock()
		}
	}()

	return candles
}
