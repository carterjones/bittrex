package main

import (
	"fmt"
	"log"
	"time"

	"github.com/carterjones/bittrex"
)

func main() {
	c := bittrex.New("", "")

	// Create a candle handler that prints candles.
	candleHandler := func(c bittrex.Candle) { fmt.Println(c) }

	// Begin processing candles at a one minute interval.
	c.ProcessCandles(1*time.Minute, candleHandler)

	// Subscribe to all Bittrex markets.
	subscribeToAllMarkets(c)

	// Wait indefinitely.
	select {}
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
