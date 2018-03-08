package bittrex

import (
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
)

// RequestParamsRecorder provides a way to record request parameters.
type RequestParamsRecorder struct {
	Params url.Values
}

// SaveParams saves the parameters passed in by the specified HTTP request.
func (rr *RequestParamsRecorder) SaveParams(r *http.Request) {
	rr.Params = r.URL.Query()
}

func mustReadTestFixture(fixture string) []byte {
	data, err := Asset("test-fixtures/" + fixture)
	panicIfErr(err)
	return data
}

func panicIfErr(err error) {
	if err != nil {
		panic(err)
	}
}

var (
	fixtureAccountGetbalances     = mustReadTestFixture("account_getbalances.json")
	fixtureAccountGetorderhistory = mustReadTestFixture("account_getorderhistory.json")
	fixturePubGetticks            = mustReadTestFixture("pub_market_getticks.json")
	fixturePublicGetmarkets       = mustReadTestFixture("public_getmarkets.json")
)

// NewMockRestServer returns a new httptest server that can be used to record
// parameters passed in by each request.
func NewMockRestServer() (*httptest.Server, *RequestParamsRecorder) {
	rr := new(RequestParamsRecorder)
	return httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rr.SaveParams(r)

		switch r.URL.Path {
		case "/api/v1.1/public/getmarkets":
			_, err := w.Write(fixturePublicGetmarkets)
			panicIfErr(err)
		case "/api/v1.1/account/getorderhistory":
			_, err := w.Write(fixtureAccountGetorderhistory)
			panicIfErr(err)
		case "/api/v1.1/account/getbalances":
			_, err := w.Write(fixtureAccountGetbalances)
			panicIfErr(err)
		case "/api/v1.1/market/cancel":
			// Note: the "market/cancel" documentation doesn't make sense at
			// https://bittrex.com/home/api
			// So this might be wrong somehow.
			_, err := w.Write([]byte(`{"success":true,"message":"","result":"my-uuid"}`))
			panicIfErr(err)
		case "/api/v1.1/market/buylimit":
			_, err := w.Write([]byte(`{"success":true,"message":"","result":{"uuid":"e606d53c-8d70-11e3-94b5-425861b86ab6"}}`))
			panicIfErr(err)
		case "/api/v1.1/market/selllimit":
			_, err := w.Write([]byte(`{"success":true,"message":"","result":{"uuid":"614c34e4-8d71-11e3-94b5-425861b86ab6"}}`))
			panicIfErr(err)
		case "/Api/v2.0/pub/market/GetTicks":
			_, err := w.Write(fixturePubGetticks)
			panicIfErr(err)
		default:
			log.Println(r.URL.Path)
		}
	})), rr
}
