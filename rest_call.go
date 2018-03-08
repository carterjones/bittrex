package bittrex

import (
	"crypto/hmac"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/pkg/errors"
)

type restCall struct {
	req     *http.Request
	res     *restResponse
	resBody []byte

	httpClient    *http.Client
	httpTransport *http.Transport

	apiKey    string
	apiSecret string

	// The host address of the Bittrex service. We split this out to simplify
	// testing.
	hostAddr string

	params map[string]string

	// A function that returns a value representing "now". We can replace this
	// value to simplify testing.
	nower nower
}

func (rc *restCall) addAuthData() {
	secret := []byte(rc.apiSecret)
	nonce := strconv.FormatInt(rc.nower.Now().Unix(), 10)

	// Extract the existing query.
	q := rc.req.URL.Query()

	// Add the API key and nonce.
	q.Add("apiKey", rc.apiKey)
	q.Add("nonce", nonce)

	// Save the new query to the request object.
	rc.req.URL.RawQuery = q.Encode()

	// Calculate the signature.
	mac := hmac.New(sha512.New, secret)
	// We ignore the error because the documentation states errors will never be
	// returned: https://golang.org/pkg/hash/#Hash
	_, _ = mac.Write([]byte(rc.req.URL.String())) // nolint: gas
	sign := mac.Sum(nil)
	signHex := hex.EncodeToString(sign)

	// Add the header to the request.
	rc.req.Header.Add("apisign", signHex)
}

func (rc *restCall) doV1_1(api string, requiresAuth bool) error {
	uri := rc.hostAddr + "/api/v1.1/" + api

	err := rc.doGenericCall(uri, requiresAuth)
	if err != nil {
		return errors.Wrap(err, "generic call failed")
	}

	return nil
}

func (rc *restCall) doV2_0(api string) error {
	uri := rc.hostAddr + "/Api/v2.0/" + api

	err := rc.doGenericCall(uri, false)
	if err != nil {
		return errors.Wrap(err, "generic call failed")
	}

	return nil
}

type restResponse struct {
	Success bool             `json:"success"`
	Message string           `json:"message"`
	Result  *json.RawMessage `json:"result"`
}

// nower is the interface that wraps Now.
//
// Now gets a time that represents the current time.
type nower interface {
	Now() time.Time
}

// defaultNower is the default value of the Connector.Nower interface, which is
// used to obtain a time that represents "now".
type defaultNower struct{}

// Now returns the result of time.Now().
func (dn defaultNower) Now() time.Time {
	return time.Now()
}

// TODO(carter): add some type of retry capability here
func (rc *restCall) doGenericCall(uri string, requiresAuth bool) error {
	// Create a client.
	var c *http.Client
	if rc.httpClient == nil {
		c = http.DefaultClient
	} else {
		c = rc.httpClient
	}

	// Adjust the transport.
	if rc.httpTransport == nil {
		c.Transport = http.DefaultTransport
		transport := c.Transport.(*http.Transport)
		transport.TLSHandshakeTimeout = 20 * time.Second
		c.Transport = transport
	} else {
		c.Transport = rc.httpTransport
	}

	// Create a new request.
	var err error
	rc.req, err = http.NewRequest("GET", uri, nil)
	if err != nil {
		return errors.Wrap(err, "get request creation failed")
	}

	// Add the URL parameters to the request.
	rc.req.URL = addParamsToURL(rc.params, rc.req.URL)

	// Prepare authorization parameters and headers.
	if requiresAuth {
		rc.addAuthData()
	}

	// Execute the API call.
	var respRaw *http.Response
	respRaw, err = c.Do(rc.req)
	if err != nil {
		return errors.Wrap(err, "failed to execute the api call")
	}

	// Get the response body.
	rc.resBody, err = getBody(respRaw.Body)
	if err != nil {
		return errors.Wrap(err, "failed to get body")
	}

	// Convert the results.
	rc.res, err = getRestResponse(rc.resBody, respRaw.Status)
	if err != nil {
		return errors.Wrap(err, "failed to get the rest response")
	}

	return nil
}

func newRestCall(httpClient *http.Client, httpTransport *http.Transport, apiKey, apiSecret, hostAddr string) *restCall {
	return &restCall{
		httpClient:    httpClient,
		httpTransport: httpTransport,
		apiKey:        apiKey,
		apiSecret:     apiSecret,
		hostAddr:      hostAddr,
		nower:         defaultNower{},
	}
}

func getBody(readCloser io.ReadCloser) (body []byte, err error) {
	defer func() {
		derr := readCloser.Close()
		if derr != nil {
			if err != nil {
				err = errors.Wrapf(err, "error in defer")
				err = errors.Wrapf(err, derr.Error())
			} else {
				err = errors.Wrap(derr, "error in defer")
			}
		}
	}()

	body, err = ioutil.ReadAll(readCloser)
	if err != nil {
		return []byte{}, errors.Wrap(err, "read failed")
	}

	return body, nil
}

func getRestResponse(body []byte, status string) (*restResponse, error) {
	// Unmarshal the response body.
	rr := &restResponse{}
	err := json.Unmarshal(body, &rr)
	if err != nil {
		debugMessage("status: %v, body: %v", status, body)
		return nil, errors.Wrap(err, "json unmarshal failed")
	}

	// Verify the response was marked as a success.
	if !rr.Success {
		return nil, errors.Errorf("success is non-existent or false: %s", string(body))
	}

	// Make sure a response was set.
	if rr.Result == nil {
		return nil, errors.Errorf("no result exists: %s", string(body))
	}

	return rr, nil
}

func addParamsToURL(params map[string]string, u *url.URL) *url.URL {
	newURL := *u

	// Extract the original query.
	q := newURL.Query()

	// Add each of the parameters.
	for k, v := range params {
		q.Add(k, v)
	}

	// Save the new query.
	newURL.RawQuery = q.Encode()

	return &newURL
}

func debugEnabled() bool {
	v := os.Getenv("DEBUG")
	return v != ""
}

func debugMessage(msg string, v ...interface{}) {
	if debugEnabled() {
		log.Printf(msg, v...)
	}
}
