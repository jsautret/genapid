package jsonrpcpredicate

import (
	"bytes"
	"crypto/subtle"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/jsautret/go-api-broker/ctx"
	"github.com/jsautret/go-api-broker/genapid"
	"github.com/jsautret/go-api-broker/internal/conf"
	"github.com/kr/pretty"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

var logLevel = zerolog.FatalLevel

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
		name         string
		conf         string
		expected     bool
		invalidParam bool // true if params values are invalid

	}{
		{
			name:         "NoConf",
			conf:         "",
			invalidParam: true,
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
  username: User1
  password: pass
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
  username: USER1
  password: passwd1
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
	zerolog.SetGlobalLevel(logLevel)
	log.Logger = zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr}).
		With().Caller().Timestamp().Logger()
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			p := New()
			cfg := getConf(t, tc.conf)
			c := ctx.New()
			init := genapid.InitPredicate(log.Logger, c, p, cfg)
			assert.Equal(t, !tc.invalidParam, init, "initPredicate")
			if init {
				assert.Equal(t,
					tc.expected, p.Call(log.Logger))
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
//gocyclo:ignore
func httpServerMock(t *testing.T) *httptest.Server {
	// Start a local HTTP server
	h := func(rw http.ResponseWriter, r *http.Request) {
		t.Logf("URL: %v", r.URL.String())
		switch r.URL.Path {
		case "/test_Empty":
			_, err := rw.Write([]byte(``))
			assert.Nil(t, err)
		case "/test_InvalidJson":
			_, err := rw.Write([]byte(`Not valid JSON`))
			assert.Nil(t, err)
		case "/test_IntegerResponse":
			_, err := rw.Write([]byte(`{"jsonrpc": "2.0", "result": 1, "id": 0}`))
			assert.Nil(t, err)

		case "/test_OneStringParam":
			type jsonResponse struct {
				Method, Jsonrpc string
				ID              int
				Params          map[string]string
			}
			response := jsonResponse{}
			if err := json.Unmarshal(streamToByte(t, r.Body),
				&response); err != nil {
				t.Fatalf("Received unvalid JSON: %v", err)
			}
			if response.Method != "test1" {
				t.Fatalf("Received method %v, expected test1", response.Method)
			}
			v := response.Params["param1"]
			if v != "value1" {
				t.Fatalf("Received param1 value %v, expected value1", v)
			}
			_, err := rw.Write([]byte(`{"jsonrpc": "2.0", "result": 1, "id": 0}`))
			assert.Nil(t, err)

		case "/test_OneIntParam":
			body := streamToByte(t, r.Body)
			type jsonResponse2 struct {
				Method, Jsonrpc string
				ID              int
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
			_, err := rw.Write([]byte(`{"jsonrpc": "2.0", "result": 1, "id": 0}`))
			assert.Nil(t, err)

		case "/test_ListParam":
			body := streamToByte(t, r.Body)
			t.Logf("%s", body)
			type jsonResponse struct {
				Method, Jsonrpc string
				ID              int
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
			_, err := rw.Write([]byte(`{"jsonrpc": "2.0", "result": 1, "id": 0}`))
			assert.Nil(t, err)

		case "/test_BasicAuth":
			user, pass, ok := r.BasicAuth()
			if !ok || subtle.ConstantTimeCompare([]byte(user),
				[]byte("user1")) != 1 ||
				subtle.ConstantTimeCompare([]byte(pass),
					[]byte("pass1")) != 1 {
				t.Fatalf("expected true, (%v,%v), got %v,(%v,%v)",
					username, password, ok, user, pass)
			}
			_, err := rw.Write([]byte(`{"jsonrpc": "2.0", "result": 1, "id": 0}`))
			assert.Nil(t, err)

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
func getConf(t *testing.T, source string) *conf.Params {
	c := conf.Params{}
	require.Nil(t,
		yaml.Unmarshal([]byte(source), &c.Conf), "YAML parsing failed")
	t.Logf("Parsed YAML:\n%# v", pretty.Formatter(c))

	return &c
}

func streamToByte(t *testing.T, stream io.Reader) []byte {
	buf := new(bytes.Buffer)
	_, err := buf.ReadFrom(stream)
	require.Nil(t, err)
	return buf.Bytes()
}
