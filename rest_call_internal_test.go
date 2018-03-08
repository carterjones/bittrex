package bittrex

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/pkg/errors"
)

func TestAddAuthData(t *testing.T) {
	cases := map[string]struct {
		apiKey    string
		apiSecret string
		exp       http.Header
	}{
		"normal": {
			apiKey:    "hello",
			apiSecret: "world",
			exp: http.Header{
				"Apisign": []string{"a276ab89467a2b549271c3c32b72006f1bad03146fe04220dde63e4844c6e9d06856b29cab9d913d1331988020653eb79e7ead408a3805c323182c85fb39a9d7"},
			},
		},
	}

	for id, tc := range cases {
		rc := newRestCall(nil, nil, tc.apiKey, tc.apiSecret, "")
		var err error
		rc.req, err = http.NewRequest("GET", "bla", nil)
		checkErr(err)

		// Use a constant time.
		rc.nower = constNower{}
		rc.addAuthData()
		equals(t, id, tc.exp, rc.req.Header)
	}
}

func TestDoV1_1_and_V2_0(t *testing.T) {
	cases := map[string]struct {
		api          string
		requiresAuth bool
		wantErr      string
	}{
		"normal": {
			api:          "my/api",
			requiresAuth: false,
		},
		"error occurred": {
			api:          "throw-error",
			requiresAuth: false,
			wantErr:      "generic call failed: failed to get the rest response",
		},
	}

	for id, tc := range cases {
		var v1_1, v2_0 bool
		done := make(chan struct{})
		testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			v1_1 = false
			v2_0 = false
			if strings.Contains(r.URL.String(), "v1.1") {
				v1_1 = true
			}
			if strings.Contains(r.URL.String(), "v2.0") {
				v2_0 = true
			}
			if v1_1 && v2_0 {
				t.Fatalf("%v: both v1.1 and v2.0 were identified in this test call's URL", id)
			}

			if strings.Contains(r.URL.String(), "throw-error") {
				// Don't return any data in the body of the response. This will
				// cause an error.
			} else {
				w.Write([]byte(`{"success":true, "result":"bla"}`))
			}
			go func() { done <- struct{}{} }()
		}))
		testClient := testServer.Client()

		rc := newRestCall(testClient, nil, "my-key", "my-secret", testServer.URL)
		rc.hostAddr = testServer.URL

		// Test v1.1 API.
		act := rc.doV1_1(tc.api, tc.requiresAuth)
		<-done
		if tc.wantErr != "" {
			errMatches(t, id, act, tc.wantErr)
		} else {
			ok(t, id, act)
			equals(t, id, true, v1_1)
		}

		// Test v2.0 API.
		act = rc.doV2_0(tc.api)
		if tc.wantErr != "" {
			errMatches(t, id, act, tc.wantErr)
		} else {
			ok(t, id, act)
			equals(t, id, true, v2_0)
		}

		testServer.Close()
	}
}

func TestDoGenericCall(t *testing.T) {
	serverResponseWriteTimeout := 500 * time.Millisecond
	successfulHandler := func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"success":true,"result":"hello world"}`))
	}
	causeWriteResponseTimeout := func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(3 * serverResponseWriteTimeout)
	}
	workingAPI := "/hello/world"

	cases := map[string]struct {
		rc           *restCall
		api          string
		customURI    string
		requiresAuth bool
		handlerFunc  http.HandlerFunc
		wantErr      string
	}{
		"normal": {
			rc:           newRestCall(nil, nil, "mykey", "mysecret", ""),
			api:          workingAPI,
			handlerFunc:  successfulHandler,
			requiresAuth: false,
		},
		"custom httpclient": {
			rc:           newRestCall(&http.Client{}, nil, "mykey", "mysecret", ""),
			api:          workingAPI,
			handlerFunc:  successfulHandler,
			requiresAuth: false,
		},
		"custom transport": {
			rc:           newRestCall(nil, &http.Transport{}, "mykey", "mysecret", ""),
			api:          workingAPI,
			handlerFunc:  successfulHandler,
			requiresAuth: false,
		},
		"fail to create get request": {
			rc:           newRestCall(nil, nil, "mykey", "mysecret", ""),
			api:          workingAPI,
			customURI:    ":",
			handlerFunc:  successfulHandler,
			requiresAuth: false,
			wantErr:      "get request creation failed: parse :: missing protocol scheme",
		},
		"failed to execute api call": {
			rc:           newRestCall(nil, nil, "mykey", "mysecret", ""),
			api:          workingAPI,
			handlerFunc:  causeWriteResponseTimeout,
			requiresAuth: false,
			wantErr:      "hello/world: EOF",
		},
		"requires auth": {
			rc:           newRestCall(nil, nil, "mykey", "mysecret", ""),
			api:          workingAPI,
			handlerFunc:  successfulHandler,
			requiresAuth: true,
		},
		"get body failure": {
			rc:  newRestCall(nil, nil, "mykey", "mysecret", ""),
			api: workingAPI,
			handlerFunc: func(w http.ResponseWriter, r *http.Request) {
				w.Write([]byte("stuff"))
				if wr, ok := w.(http.Hijacker); ok {
					conn, _, err := wr.Hijack()
					checkErr(err)
					fmt.Fprint(w, err)
					conn.Close()
				}
			},
			wantErr: "failed to get body: read failed: unexpected EOF",
		},
		"get rest response failure": {
			rc:  newRestCall(nil, nil, "mykey", "mysecret", ""),
			api: workingAPI,
			handlerFunc: func(w http.ResponseWriter, r *http.Request) {
				w.Write([]byte(`{}`))
			},
			requiresAuth: false,
			wantErr:      "failed to get the rest response: success is non-existent or false: {}",
		},
	}

	for id, tc := range cases {
		ts := httptest.NewUnstartedServer(http.HandlerFunc(tc.handlerFunc))
		ts.Config.WriteTimeout = serverResponseWriteTimeout
		ts.Start()

		var uri string
		if tc.customURI != "" {
			uri = tc.customURI
		} else {
			uri = ts.URL + tc.api
		}

		act := tc.rc.doGenericCall(uri, tc.requiresAuth)
		if tc.wantErr != "" {
			errMatches(t, id, act, tc.wantErr)
		} else {
			if tc.requiresAuth {
				if tc.rc.req.Header.Get("apisign") == "" {
					t.Errorf("%v: expected auth header, but none was set", id)
				}
			}
			ok(t, id, act)
		}
	}
}

func TestNewRestCall(t *testing.T) {
	cases := map[string]struct {
		httpClient    *http.Client
		httpTransport *http.Transport
		apiKey        string
		apiSecret     string
		hostAddr      string
		exp           *restCall
	}{
		"normal": {
			httpClient:    &http.Client{Timeout: 123 * time.Millisecond},
			httpTransport: &http.Transport{MaxIdleConns: 456},
			apiKey:        "hello",
			apiSecret:     "world",
			hostAddr:      "https://localhost",
			exp: &restCall{
				httpClient:    &http.Client{Timeout: 123 * time.Millisecond},
				httpTransport: &http.Transport{MaxIdleConns: 456},
				apiKey:        "hello",
				apiSecret:     "world",
				hostAddr:      "https://localhost",
				nower:         defaultNower{},
			},
		},
		"empty": {
			httpClient:    nil,
			httpTransport: nil,
			apiKey:        "",
			apiSecret:     "",
			hostAddr:      "",
			exp: &restCall{
				httpClient:    nil,
				httpTransport: nil,
				apiKey:        "",
				apiSecret:     "",
				hostAddr:      "",
				nower:         defaultNower{},
			},
		},
	}

	for id, tc := range cases {
		act := newRestCall(tc.httpClient, tc.httpTransport, tc.apiKey, tc.apiSecret, tc.hostAddr)
		equals(t, id, tc.exp, act)
	}
}

func TestGetBody(t *testing.T) {
	cases := map[string]struct {
		body    io.ReadCloser
		exp     []byte
		wantErr string
	}{
		"successful": {
			body: testReadCloser{
				Reader: strings.NewReader("hello world"),
			},
			exp: []byte(`hello world`),
		},
		"read failure": {
			body: testReadCloser{
				readErr: "sample 123 read 456 failure 789",
			},
			wantErr: "sample 123 read 456 failure 789",
		},
		"defer close failure": {
			body: testReadCloser{
				Reader:   strings.NewReader("hello world"),
				closeErr: "sample 789 close 654 failure 321",
			},
			wantErr: "sample 789 close 654 failure 321",
		},
		"defer close failure after read error": {
			body: testReadCloser{
				readErr:  "sample inner read error",
				closeErr: "sample outer close error",
			},
			wantErr: "sample outer close error: error in defer: read failed: sample inner read error",
		},
	}
	for id, tc := range cases {
		act, err := getBody(tc.body)
		if tc.wantErr != "" {
			errMatches(t, id, err, tc.wantErr)
		} else {
			equals(t, id, tc.exp, act)
			ok(t, id, err)
		}
	}
}

func TestGetRestResponse(t *testing.T) {
	cases := map[string]struct {
		body    []byte
		exp     *restResponse
		debug   bool
		wantErr string
	}{
		"successful": {
			body: []byte(`{"success":true,"result":"stuff and things"}`),
			exp: &restResponse{
				Success: true,
				Message: "",
				Result:  func() *json.RawMessage { msg := json.RawMessage(`"stuff and things"`); return &msg }(),
			},
		},
		"empty body": {
			body:    []byte{},
			wantErr: "json unmarshal failed: unexpected end of JSON input",
		},
		"empty body (with debug)": {
			body:    []byte{},
			wantErr: "json unmarshal failed: unexpected end of JSON input",
			debug:   true,
		},
		"empty json object": {
			body:    []byte("{}"),
			wantErr: "success is non-existent or false: {}",
		},
		"success is false": {
			body:    []byte(`{"success":false}`),
			wantErr: `success is non-existent or false: {"success":false}`,
		},
		"no result": {
			body:    []byte(`{"success":true}`),
			wantErr: `no result exists: {"success":true}`,
		},
	}

	for id, tc := range cases {
		if tc.debug {
			os.Setenv("DEBUG", "true")
		}

		act, err := getRestResponse(tc.body, "example status")
		if tc.wantErr != "" {
			errMatches(t, id, err, tc.wantErr)
		} else {
			equals(t, id, tc.exp, act)
			ok(t, id, err)
		}

		if tc.debug {
			os.Unsetenv("DEBUG")
		}
	}
}

func TestAddParamsToURL(t *testing.T) {
	cases := map[string]struct {
		params map[string]string
		u      string
		exp    string
	}{
		"normal": {
			params: map[string]string{"hello": "world"},
			u:      "http://localhost/blabla",
			exp:    "http://localhost/blabla?hello=world",
		},
		"no params": {
			params: map[string]string{},
			u:      "http://localhost/blabla",
			exp:    "http://localhost/blabla",
		},
		"multiple params": {
			params: map[string]string{
				"hello1": "world",
				"hello2": "world",
			},
			u:   "http://localhost/blabla",
			exp: "http://localhost/blabla?hello1=world&hello2=world",
		},
	}

	for id, tc := range cases {
		u, err := url.Parse(tc.u)
		checkErr(err)
		exp, err := url.Parse(tc.exp)
		checkErr(err)

		act := addParamsToURL(tc.params, u)
		equals(t, id, exp, act)
	}
}

type testReadCloser struct {
	*strings.Reader
	readErr  string
	closeErr string
}

func (rc testReadCloser) Read(p []byte) (int, error) {
	if rc.readErr != "" {
		return 0, errors.New(rc.readErr)
	}
	return rc.Reader.Read(p)
}

func (rc testReadCloser) Close() error {
	if rc.closeErr != "" {
		return errors.New(rc.closeErr)
	}
	return nil
}

type constNower struct{}

func (constNower) Now() time.Time {
	return time.Date(2010, 7, 5, 4, 3, 2, 1, time.UTC)
}

func checkErr(err error) {
	if err != nil {
		fmt.Printf("%+v\n", err)
		os.Exit(1)
	}
}
