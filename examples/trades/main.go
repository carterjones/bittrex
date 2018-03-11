package main

import (
	"log"
	"strings"

	"github.com/carterjones/bittrex"
)

func main() {
	c := bittrex.New("", "")

	// Register a handler to print each trade.
	printTrade := func(t bittrex.Trade) { log.Println(t) }
	c.Register(printTrade)

	// Get all the markets on Bittrex.
	log.Println("Retrieving all markets...")
	ms, err := c.Markets()
	panicIfErr(err)

	// Print the market names.
	var marketNames []string
	for _, m := range ms {
		marketNames = append(marketNames, m.Name)
	}
	log.Println("Markets:", strings.Join(marketNames, ", "))

	// Subscribe to each market.
	log.Println("Subscribing to markets. This requires bypassing CloudFlare, which usually takes 5-10 seconds.")
	for _, m := range ms {
		err = c.Subscribe(m.Name, panicIfErr)
		panicIfErr(err)
	}
	log.Println("Successfully subscribed to all markets.")

	// Wait indefinitely.
	select {}
}

func panicIfErr(err error) {
	if err != nil {
		log.Panic(err)
	}
}
