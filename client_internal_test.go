package bittrex

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"

	cfscraper "github.com/carterjones/go-cloudflare-scraper"
	"github.com/carterjones/signalr"
)

func TestClient_SetCustomID(t *testing.T) {
	cases := map[string]struct {
		client  *Client
		id      string
		wantErr string
	}{
		"normal": {
			client: New("", ""),
			id:     "hello world",
		},
		"error": {
			client:  &Client{},
			id:      "hello world",
			wantErr: "bittrex client has not been initialized",
		},
	}

	for id, tc := range cases {
		act := tc.client.SetCustomID(tc.id)
		if tc.wantErr != "" {
			errMatches(t, id, act, tc.wantErr)
		} else {
			equals(t, id, tc.id, tc.client.signalrC.CustomID)
			ok(t, id, act)
		}
	}
}

func TestClient_Subscribe(t *testing.T) {
	cases := map[string]struct {
		client   *Client
		serverFn http.HandlerFunc
		wantErr  string
	}{
		"normal": {
			client:   New("my-key", "my-secret"),
			serverFn: signalr.TestCompleteHandler,
		},
		"initialization failure": {
			client:   New("my-key", "my-secret"),
			serverFn: func(w http.ResponseWriter, r *http.Request) {},
			wantErr:  "failed to start the underlying SignalR client",
		},
	}

	for id, tc := range cases {
		ts := httptest.NewUnstartedServer(http.HandlerFunc(tc.serverFn))
		ts.Start()

		tc.client.HTTPClient = ts.Client()
		tc.client.HostAddr = ts.URL
		tc.client.signalrC.Host = strings.Replace(ts.URL, "http://", "", -1)
		tc.client.signalrC.Scheme = signalr.HTTP

		err := tc.client.Subscribe("BTC-LTC", func(error) {})
		if tc.wantErr != "" {
			errMatches(t, id, err, tc.wantErr)
			equals(t, id, false, tc.client.started)
		} else {
			equals(t, id, true, tc.client.started)
			ok(t, id, err)
		}
	}
}

func TestNewUnstarted(t *testing.T) {
	cases := map[string]struct {
		apiKey    string
		apiSecret string
		exp       *Client
	}{
		"normal": {
			apiKey:    "my-key",
			apiSecret: "my-secret",
			exp: &Client{
				APIKey:    "my-key",
				APISecret: "my-secret",
				HostAddr:  "https://bittrex.com",
			},
		},
	}

	for id, tc := range cases {
		act := New(tc.apiKey, tc.apiSecret)
		equalsClient(t, id, act, tc.exp)
	}
}

func equalsClient(t *testing.T, id string, c1 *Client, c2 *Client) {
	// Verify the base settings of the SignalR client were set.
	equals(t, id, c1.signalrC.Host, "socket.bittrex.com")
	equals(t, id, c1.signalrC.Protocol, "1.5")
	equals(t, id, c1.signalrC.Endpoint, "/signalr")
	equals(t, id, c1.signalrC.ConnectionData, `[{"name":"corehub"}]`)

	// Verify the transport was set to the CloudFlare scraper transport.
	if _, ok := c1.signalrC.HTTPClient.Transport.(*cfscraper.Transport); !ok {
		t.Errorf("%v: expected CloudFlare scraper transport, but found transport of type %v",
			id, reflect.TypeOf(c1.signalrC.HTTPClient.Transport))
	}

	// Verify the cookie jar was set correctly.
	transport := c1.signalrC.HTTPClient.Transport.(*cfscraper.Transport)
	equals(t, id, c1.signalrC.HTTPClient.Jar, transport.Cookies)

	// Verify the headers were set properly.
	equals(t, id, c1.signalrC.Headers["User-Agent"], "Mozilla/5.0 (Windows NT 6.1) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/41.0.2228.0 Safari/537.36")

	// Verify the retry wait duration default was set.
	equals(t, id, c1.signalrC.RetryWaitDuration, 10*time.Second)

	// Zero out the SignalR client and compare the rest of the Bittrex
	// client structure.
	c1.signalrC = nil
	c2.signalrC = nil
	equals(t, id, c2, c1)
}
