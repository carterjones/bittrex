package bittrex

//go:generate go-bindata -pkg internal -nometadata -o ./internal/bindata.go test-fixtures
//go:generate go fmt ./internal/bindata.go

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/carterjones/signalr"
	"github.com/carterjones/signalr/hubs"
	"github.com/pkg/errors"
)

// Client provides an interface by which people can call the Bittrex API
// endpoints (both the REST and the WebSocket APIs).
type Client struct {
	// HTTPClient is optional. It can be used to inject a custom underlying HTTP
	// client to perform HTTP operations.
	HTTPClient *http.Client

	// HTTPTransport is optional. It can be used to inject a custom underlying
	// HTTP transport into the underlying HTTP client.
	HTTPTransport *http.Transport

	// MaxRetries indicates the maximum number of times to attempt an HTTP
	// operation before failing.
	MaxRetries int

	// RetryWaitDuration is used to determine how long to wait between retries
	// of various operations.
	RetryWaitDuration time.Duration

	// A Bittrex-supplied API key and secret.
	APIKey    string
	APISecret string

	// The underlying SignalR client.
	signalrC *signalr.Client

	// The current message ID that is sent to the SignalR client with each
	// message.
	currentMsgID int

	// Started indicates if the underlying SignalR client has been started.
	started    bool
	startedMux sync.Mutex

	// tradeHandlers holds all of the registered trade handler functions.
	tradeHandlers    []TradeHandler
	tradeHandlersMux sync.Mutex

	// HostAddr address is the address of the Bittrex server providing the service
	// we're using. Using this variable allows us to more simply test
	// functionality in this package.
	HostAddr string
}

// Markets gets the markets that are traded on Bittrex.
func (c *Client) Markets() ([]Market, error) {
	rc := newRestCall(c.HTTPClient, c.HTTPTransport, c.APIKey, c.APISecret, c.HostAddr)
	err := rc.doV1_1("public/getmarkets", false)
	if err != nil {
		err = errors.Wrap(err, "public/getmarkets failed")
		return []Market{}, err
	}

	var ms []Market
	err = json.Unmarshal(*rc.res.Result, &ms)
	if err != nil {
		return []Market{}, errors.Wrap(err, "json unmarshal failed")
	}

	return ms, nil
}

// Balances gets the balances held for all currencies in the account.
func (c *Client) Balances() ([]Balance, error) {
	rc := newRestCall(c.HTTPClient, c.HTTPTransport, c.APIKey, c.APISecret, c.HostAddr)
	err := rc.doV1_1("account/getbalances", true)
	if err != nil {
		return []Balance{}, errors.Wrap(err, "account/getbalances failed")
	}

	var bs []Balance
	err = json.Unmarshal(*rc.res.Result, &bs)
	if err != nil {
		return []Balance{}, errors.Wrap(err, "json unmarshal failed")
	}

	return bs, nil
}

// OrderHistory gets the latest orders made through Bittrex for the user's
// account.
func (c *Client) OrderHistory() ([]Order, error) {
	rc := newRestCall(c.HTTPClient, c.HTTPTransport, c.APIKey, c.APISecret, c.HostAddr)
	err := rc.doV1_1("account/getorderhistory", true)
	if err != nil {
		return []Order{}, errors.Wrap(err, "account/getorderhistory failed")
	}

	var orders []Order
	err = json.Unmarshal(*rc.res.Result, &orders)
	if err != nil {
		return []Order{}, errors.Wrap(err, "json unmarshal failed")
	}

	return orders, nil
}

// LimitSell sends a request to create a limit sell order.
func (c *Client) LimitSell(market string, quantity, rate float64) error {
	rc := newRestCall(c.HTTPClient, c.HTTPTransport, c.APIKey, c.APISecret, c.HostAddr)
	api := "market/selllimit"

	// Prepare the parameters.
	rc.params = map[string]string{
		"market":   market,
		"quantity": strconv.FormatFloat(quantity, 'f', 8, 64),
		"rate":     strconv.FormatFloat(rate, 'f', 8, 64),
	}

	// If the result is successful, no error will be thrown. The only bit of
	// information returned is the uuid of the sell order, which we don't care
	// about. Therefore we ignore the response.
	err := rc.doV1_1(api, true)
	if err != nil {
		return errors.Wrap(err, "market/selllimit failed")
	}

	return nil
}

// LimitBuy sends a request to create a limit buy order.
func (c *Client) LimitBuy(market string, quantity, rate float64) error {
	rc := newRestCall(c.HTTPClient, c.HTTPTransport, c.APIKey, c.APISecret, c.HostAddr)
	api := "market/buylimit"

	// Prepare the parameters.
	rc.params = map[string]string{
		"market":   market,
		"quantity": strconv.FormatFloat(quantity, 'f', 8, 64),
		"rate":     strconv.FormatFloat(rate, 'f', 8, 64),
	}

	// If the result is successful, no error will be thrown. The only bit of
	// information returned is the uuid of the sell order, which we don't care
	// about. Therefore we ignore the response.
	err := rc.doV1_1(api, true)
	if err != nil {
		return errors.Wrap(err, "market/buylimit failed")
	}

	return nil
}

// Cancel sends a request to cancel an order with the specified UUID.
func (c *Client) Cancel(orderUUID string) error {
	rc := newRestCall(c.HTTPClient, c.HTTPTransport, c.APIKey, c.APISecret, c.HostAddr)
	api := "market/cancel"

	// Prepare the parameters.
	rc.params = map[string]string{
		"uuid": orderUUID,
	}

	// If the result is successful, no error will be thrown. The only bit of
	// information returned is the uuid of the sell order, which we don't care
	// about. Therefore we ignore the response.
	err := rc.doV1_1(api, true)
	if err != nil {
		return errors.Wrap(err, "market/cancel failed")
	}

	return nil
}

// Subscribe sends a request to Bittrex to start sending us the market data for
// the indicated market.
func (c *Client) Subscribe(market string, errHandler ErrHandler) error {
	err := c.websocketReady(errHandler)
	if err != nil {
		return errors.Wrap(err, "underlying signalr client is not ready")
	}

	msgs := []interface{}{market}
	msgID := c.currentMsgID
	c.currentMsgID++
	hcm := hubs.ClientMsg{
		H: "corehub",
		M: "SubscribeToExchangeDeltas",
		A: msgs,
		I: msgID,
	}

	err = c.signalrC.Send(hcm)
	if err != nil {
		return errors.Wrap(err, "failed to send via signalr client")
	}

	return nil
}

// Register saves the specified trade handler to a slice of handlers that will
// be run against each incoming trade.
func (c *Client) Register(h TradeHandler) {
	c.tradeHandlersMux.Lock()
	defer c.tradeHandlersMux.Unlock()
	c.tradeHandlers = append(c.tradeHandlers, h)
}

func (c *Client) websocketReady(errHandler ErrHandler) error {
	// Return if no SignalR client exists.
	if c.signalrC == nil {
		return errors.New("underlying signalr client is not initialized")
	}

	// Protect the started flag.
	c.startedMux.Lock()
	defer c.startedMux.Unlock()

	// Return if the client has already been started.
	if c.started {
		return nil
	}

	// Prepare a message channel.
	msgs := make(chan signalr.Message)

	// Initialize the SignalR client.
	msgHandler := func(msg signalr.Message) { msgs <- msg }
	err := c.signalrC.Run(msgHandler, signalr.ErrHandler(errHandler))
	if err != nil {
		return errors.Wrap(err, "failed to start the underlying SignalR client")
	}

	// Process the messages.
	go c.processMessages(msgs, errHandler)

	c.started = true

	return nil
}

// Ticks gets a little under 10 days of candle data (14365 minutes) at the one
// minute interval. However, it is usually missing the last few minutes, so a
// way to monitor trades (such as the Trades() function) is still required to
// get up-to-the-minute data.
//
// As a general approach when monitoring data for strategic analysis, this
// function should be called once a minute for a few minutes until the
// timestamps from the live trade data (Trades() function) overlap with the data
// returned by this function. You shouldn't have to wait more than about 5-10
// minutes for this to occur; at that point you likely only need to use live
// trade data and perhaps then use this function as the source of truth for
// validating your data.
//
// Intervals provided by the underlying Bittrex API are: day, hour, thirtyMin,
// fiveMin, and oneMin.
func (c *Client) Ticks(market string) ([]Tick, error) {
	rc := newRestCall(c.HTTPClient, c.HTTPTransport, c.APIKey, c.APISecret, c.HostAddr)
	interval := "oneMin"

	// Set the parameters.
	rc.params = map[string]string{
		"marketName":   market,
		"tickInterval": interval,
	}

	// Perform the API call.
	err := rc.doV2_0("pub/market/GetTicks")
	if err != nil {
		return []Tick{}, errors.Wrap(err, "pub/market/GetTicks failed")
	}

	// Extract the result into a usable variable.
	result := rc.res.Result

	// Return if no result was returned. This can happen even in successful
	// situations. For example, the JSON response may look like this:
	//
	// {"success":true,"message":"","result":null}
	if result == nil {
		return []Tick{}, nil
	}

	// Convert the results.
	var ts []Tick
	err = json.Unmarshal(*result, &ts)
	if err != nil {
		return []Tick{}, errors.Wrap(err, "json unmarshal failed")
	}

	return ts, nil
}

// SetCustomID assigns the specified id to the underlying SignalR client.
func (c *Client) SetCustomID(id string) error {
	if c.signalrC == nil {
		return errors.New("bittrex client has not been initialized")
	}

	c.signalrC.CustomID = id
	return nil
}

// ProcessCandles monitors the trade data for all subscribed markets and
// produces candle data for the specified interval.
func (c *Client) ProcessCandles(interval time.Duration, candleHandler CandleHandler) {
	// Register a handler to funnel each trade to a trades channel.
	trades := make(chan Trade)
	c.Register(func(t Trade) { trades <- t })

	// Create a holding place for the candles for this interval.
	candles := make(map[string]Candle)
	candlesMux := sync.Mutex{}

	// Start a goroutine that updates candle values as each trade comes in.
	go func() {
		for {
			t := <-trades

			candlesMux.Lock()

			var candle Candle
			var ok bool
			if candle, ok = candles[t.Market()]; !ok {
				// If the candle does not already exist, then create one.
				candles[t.Market()] = Candle{
					Market: t.Market(),
					Open:   t.Price,
					High:   t.Price,
					Low:    t.Price,
					Close:  t.Price,
					Volume: t.Quantity,
				}

				// Move to the next trade.
				candlesMux.Unlock()
				continue
			}

			// If the candle does exist, then modify it as necessary.

			// Set the open value if none is set.
			if candle.Open == 0.0 {
				candle.Open = t.Price
			}

			// Set the high value.
			if t.Price > candle.High {
				candle.High = t.Price
			}

			// Set the low value.
			if t.Price < candle.Low {
				candle.Low = t.Price
			}

			// Set the last price. This will eventually be used as the close
			// value.
			candle.Close = t.Price

			// Increase the volume.
			candle.Volume += t.Quantity

			// Save the new value.
			candles[t.Market()] = candle

			candlesMux.Unlock()
		}
	}()

	// Start a goroutine that waits for the specified interval, prints the
	// current candle values, and then resets the candle map.
	go func() {
		for {
			// Wait for the specified interval.
			<-time.After(interval)

			// Save the current time.
			now := time.Now()

			// Lock the map.
			candlesMux.Lock()

			// Iterate over the candles for this interval.
			for k, v := range candles {
				// Set the time to now minus the specified interval. This
				// represents when the interval started.
				v.Time = now.Add(-1 * interval)

				// Process the candle.
				go candleHandler(v)

				// Update the candle values so they carry over data to the next
				// interval.
				v.Open = v.Close
				v.High = v.Close
				v.Low = v.Close
				candles[k] = v
			}

			// Unlock the map.
			candlesMux.Unlock()
		}
	}()
}

// New creates a new Bittrex client.
func New(apiKey, apiSecret string) *Client {
	c := new(Client)

	// Set the API key and secret.
	c.APIKey = apiKey
	c.APISecret = apiSecret

	// Set the default host address.
	c.HostAddr = "https://bittrex.com"

	// Set up the underlying SignalR client.
	signalrC := signalr.New(
		"socket.bittrex.com",
		"1.5",
		"/signalr",
		`[{"name":"c2"}]`,
		nil,
	)

	// Set the user agent to one that looks like a browser.
	signalrC.Headers["User-Agent"] = "Mozilla/5.0 (Windows NT 6.1) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/41.0.2228.0 Safari/537.36"

	// Set the retry duration default.
	signalrC.RetryWaitDuration = 10 * time.Second

	// Save the SignalR client.
	c.signalrC = signalrC

	return c
}

// TradeHandler processes a trade.
type TradeHandler func(t Trade)

// CandleHandler processes a candle.
type CandleHandler func(c Candle)

// ErrHandler processes an error.
type ErrHandler func(err error)

// This processes SignalR messages.
func (c *Client) processMessages(msgs chan signalr.Message, errHandler ErrHandler) {
	for msg := range msgs {
		if !c.processMessage(msg, errHandler) {
			return
		}
	}
}

// Process a single SignalR message.
func (c *Client) processMessage(msg signalr.Message, errHandler ErrHandler) bool {
	// Assume that the message is successfully processed. Prove otherwise.
	ok := true

	// Within each SignalR message is a slice of Bittrex messages.
	for _, bittrexMsg := range msg.M {
		// Verify the Bittrex message type is updateExchangeState.
		if bittrexMsg.M != "updateExchangeState" {
			continue
		}

		// Process each of the arguments.
		for _, arg := range bittrexMsg.A {
			ok = c.processBittrexMsgArg(arg, errHandler)
			if !ok {
				return false
			}
		}
	}

	return ok
}

// Process a single argument from a Bittrex message.
func (c *Client) processBittrexMsgArg(arg interface{}, errHandler ErrHandler) bool {
	data, err := json.Marshal(arg)
	if err != nil {
		go errHandler(errors.Wrap(err, "json marshal failed"))
		return false
	}

	var eu exchangeUpdate
	err = json.Unmarshal(data, &eu)
	if err != nil {
		go errHandler(errors.Wrap(err, "json unmarshal failed"))
		return false
	}

	for _, t := range eu.Fills {
		marketParts := strings.Split(eu.MarketName, "-")
		bc := marketParts[0]
		mc := marketParts[1]
		tType := InvalidTradeType

		switch t.OrderType {
		case "BUY":
			tType = BuyType
		case "SELL":
			tType = SellType
		default:
			go errHandler(errors.Errorf("invalid trade type: %v", t.OrderType))
			return false
		}

		// Parse the time.
		var parsedTime time.Time
		parsedTime, err = t.time()
		if err != nil {
			go errHandler(errors.Wrap(err, "time parse error"))
			return false
		}

		// Create a new trade.
		t := Trade{
			BaseCurrency:   bc,
			MarketCurrency: mc,
			Type:           tType,
			Price:          t.Rate,
			Quantity:       t.Quantity,
			Time:           parsedTime,
		}

		// Process the trade using the trade handlers.
		c.tradeHandlersMux.Lock()
		for _, h := range c.tradeHandlers {
			go h(t)
		}
		c.tradeHandlersMux.Unlock()
	}

	// If we made it to this point, then the argument was successfully
	// processed.
	return true
}
