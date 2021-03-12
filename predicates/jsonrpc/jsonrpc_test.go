package jsonrpc

import (
	"bytes"
	"crypto/subtle"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/jsautret/go-api-broker/context"
	"github.com/jsautret/go-api-broker/internal/conf"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v2"
)

var logLevel = zerolog.FatalLevel

//var logLevel = zerolog.DebugLevel

var (
	username = "user1"
	password = "pass1"
)

func TestJsonrpc(t *testing.T) {
	// Start a local HTTP server
	server := httpServerMock(t)
	// Close the server when test finishes
	defer server.Close()

	cases := []struct {
		name     string
		conf     string
		expected bool
	}{
		{
			name:     "NoConf",
			conf:     "",
			expected: false,
		},
		{
			name: "EmptyResponse",
			conf: `
url: ` + server.URL + `/test_Empty
procedure: test1
`,
			expected: false,
		},
		{
			name: "invalidJson",
			conf: `
url: ` + server.URL + `/test_InvalidJson
procedure: test1
`,
			expected: false,
		},
		{
			name: "ValidResponse",
			conf: `
url: ` + server.URL + `/test_IntegerResponse
procedure: test1
`,
			expected: true,
		},
		{
			name: "OneStringParam",
			conf: `
url: ` + server.URL + `/test_OneStringParam
procedure: test1
params:
  param1: value1
basic_auth:
  username: kodi
  password: ghghpczyuq
`,
			expected: true,
		},
		{
			name: "OneIntParam",
			conf: `
url: ` + server.URL + `/test_OneIntParam
procedure: test1
params: 42
basic_auth:
  username: kodi
  password: ghghpczyuq
`,
			expected: true,
		},
		{
			name: "ListParam",
			conf: `
url: ` + server.URL + `/test_ListParam
procedure: test1
params:
  - 8
  - value2
basic_auth:
  username: kodi
  password: ghghpczyuq
`,
			expected: true,
		},
		{
			name: "BasicAuth",
			conf: `
url: ` + server.URL + `/test_BasicAuth
procedure: test1
basic_auth:
  username: ` + username + `
  password: ` + password + `
`,
			expected: true,
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			conf := getConf(t, c.conf)
			ctx := context.Ctx{
				R:       make(map[string]map[string]interface{}),
				Results: make(map[string]interface{}),
			}
			if r := Call(&ctx, conf); r != c.expected {
				t.Errorf("Should have returned %v, got %v",
					c.expected, r)
			}
		})
	}
}

func TestMain(m *testing.M) {
	zerolog.SetGlobalLevel(logLevel)
	log.Logger = zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr}).
		With().Timestamp().Logger()
	os.Exit(m.Run())
}

/***************************************************************************
  JSONRCP server mock
  ***************************************************************************/
func httpServerMock(t *testing.T) *httptest.Server {
	// Start a local HTTP server
	h := func(rw http.ResponseWriter, r *http.Request) {
		t.Logf("URL: %v", r.URL.String())
		switch r.URL.Path {
		case "/test_Empty":
			rw.Write([]byte(``))
		case "/test_InvalidJson":
			rw.Write([]byte(`Not valid JSON`))
		case "/test_IntegerResponse":
			rw.Write([]byte(`{"jsonrpc": "2.0", "result": 1, "id": 0}`))

		case "/test_OneStringParam":
			type jsonResponse struct {
				Method, Jsonrpc string
				Id              int
				Params          map[string]string
			}
			response := jsonResponse{}
			if err := json.Unmarshal(streamToByte(r.Body), &response); err != nil {
				t.Fatalf("Received unvalid JSON: %v", err)
			}
			if response.Method != "test1" {
				t.Fatalf("Received method %v, expected test1", response.Method)
			}
			v := response.Params["param1"]
			if v != "value1" {
				t.Fatalf("Received param1 value %v, expected value1", v)
			}
			rw.Write([]byte(`{"jsonrpc": "2.0", "result": 1, "id": 0}`))

		case "/test_OneIntParam":
			body := streamToByte(r.Body)
			type jsonResponse2 struct {
				Method, Jsonrpc string
				Id              int
				Params          []int
			}
			response := jsonResponse2{}
			if err := json.Unmarshal(body, &response); err != nil {
				t.Fatalf("Received unvalid JSON: %v", err)
			}
			if response.Method != "test1" {
				t.Fatalf("Received method %v, expected test1",
					response.Method)
			}
			v := response.Params

			if len(v) != 1 || v[0] != 42 {
				t.Fatalf("Received param1 value %v, expected [42]", v)
			}
			rw.Write([]byte(`{"jsonrpc": "2.0", "result": 1, "id": 0}`))

		case "/test_ListParam":
			body := streamToByte(r.Body)
			t.Logf("%s", body)
			type jsonResponse struct {
				Method, Jsonrpc string
				Id              int
				Params          []interface{}
			}
			response := jsonResponse{}
			if err := json.Unmarshal(body, &response); err != nil {
				t.Fatalf("Received unvalid JSON: %v", err)
			}
			v := response.Params
			if int(v[0].(float64)) != 8 || v[1].(string) != "value2" {
				t.Fatalf("Received value %v, expected [8, \"value2\"]", v)
			}
			rw.Write([]byte(`{"jsonrpc": "2.0", "result": 1, "id": 0}`))

		case "/test_BasicAuth":
			user, pass, ok := r.BasicAuth()
			if !ok || subtle.ConstantTimeCompare([]byte(user),
				[]byte("user1")) != 1 ||
				subtle.ConstantTimeCompare([]byte(pass),
					[]byte("pass1")) != 1 {
				t.Fatalf("expected true, (%v,%v), got %v,(%v,%v)",
					username, password, ok, user, pass)
			}
			rw.Write([]byte(`{"jsonrpc": "2.0", "result": 1, "id": 0}`))
		default:
			t.Fatalf("Unexpected path: %v", r.URL.Path)
			rw.WriteHeader(http.StatusInternalServerError)
		}
	}

	return httptest.NewServer(
		http.HandlerFunc(h))
}

/***************************************************************************
  Helpers
  ***************************************************************************/
func getConf(t *testing.T, source string) conf.Params {
	c := conf.Params{}
	if err := yaml.Unmarshal([]byte(source), &c); err != nil {
		t.Errorf("Should not have returned parsing error")
	}
	return c
}

func getConfB(source string) conf.Params {
	c := conf.Params{}
	yaml.Unmarshal([]byte(source), &c)
	return c
}

func streamToByte(stream io.Reader) []byte {
	buf := new(bytes.Buffer)
	buf.ReadFrom(stream)
	return buf.Bytes()
}
