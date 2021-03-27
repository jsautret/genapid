package httppredicate

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/jsautret/genapid/app/conf"
	"github.com/jsautret/genapid/ctx"
	"github.com/jsautret/genapid/genapid"
	"github.com/kr/pretty"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

var logLevel = zerolog.FatalLevel

var (
	username = "user1"
	password = "pass1"
)

type ctrl struct {
	t         *testing.T
	path      string            // expected URL path
	method    string            // expected HTTP method
	headers   map[string]string // expected HTTP headers
	params    map[string]string // expected URL query params
	body      string            // expected HTTP body
	ct        string            // expected content-type
	basicAuth []string          // expected auth
	resp      string            // content returned by the mock API
}

func TestHTTP(t *testing.T) {
	cases := []struct {
		name         string
		conf         string
		expected     bool        // return of predicate
		ctrl         ctrl        // passed to mock API
		invalidParam bool        // true if params values are invalid
		expRes       interface{} // expected predicate result
	}{
		{
			name:         "NoConf",
			conf:         "",
			invalidParam: true,
		},
		{
			name: "badUrl",
			conf: `
url: test/bad
`,
			invalidParam: true,
		},
		{
			name: "badMethod",
			conf: `
url: http://test/badmethod
method: BAD
`,
			invalidParam: true,
		},
		{
			name: "badBody",
			conf: `
url: http://test/badbody
body:
  json:
    k: v
  string: content
`,
			invalidParam: true,
		},
		{
			name:     "postNobody",
			expected: true,
			expRes:   "",
			ctrl: ctrl{
				path:   "/postnobody",
				method: "POST",
			},
			conf: `
url: http://test/postnobody
method: post
`,
		},
		{
			name:     "GetBody",
			expected: true,
			expRes:   "",
			ctrl: ctrl{
				path:   "/getbody",
				method: "GET",
			},
			conf: `
url: http://test/getbody
method: get
body:
  string: getbody
`,
		},
		{
			name:     "SimpleGet",
			expected: true,
			expRes:   "",
			ctrl: ctrl{
				path:   "/get",
				method: "GET",
			},
			conf: `
url: http://test/get
`,
		},
		{
			name:     "GetUrlParams",
			expected: true,
			expRes:   "",
			ctrl: ctrl{
				path:   "/params",
				method: "GET",
				params: map[string]string{"k1": "v1", "k2": "v2"},
			},
			conf: `
url: http://test/params?k1=v1&k2=v2
`,
		},
		{
			name:     "GetParams",
			expected: true,
			expRes:   "",
			ctrl: ctrl{
				path:   "/params",
				method: "GET",
				params: map[string]string{"k1": "v1", "k2": "v2"},
			},
			conf: `
url: http://test/params
params:
  k1: v1
  k2: v2
`,
		},
		{
			name:     "GetHeaders",
			expected: true,
			expRes:   "",
			ctrl: ctrl{
				path:    "/headers",
				method:  "GET",
				headers: map[string]string{"h1": "v1", "h2": "v2"},
			},
			conf: `
url: http://test/headers
headers:
  h1: v1
  h2: v2
`,
		},
		{
			name:     "GetParamsHeaders",
			expected: true,
			expRes:   "",
			ctrl: ctrl{
				path:    "/ph",
				method:  "GET",
				headers: map[string]string{"h1": "v1", "h2": "v2"},
				params:  map[string]string{"k1": "v1", "k2": "v2"},
			},
			conf: `
url: http://test/ph
params:
  k1: v1
  k2: v2
headers:
  h1: v1
  h2: v2
`,
		},
		{
			name:     "GetAuth",
			expected: true,
			expRes:   "",
			ctrl: ctrl{
				path:      "/auth",
				method:    "GET",
				headers:   map[string]string{"h1": "v1", "h2": "v2"},
				basicAuth: []string{"myuser", "mypasswd"},
			},
			conf: `
url: http://test/auth
headers:
  h1: v1
  h2: v2
basic_auth:
  username: myuser
  password: mypasswd
`,
		},
		{
			name:     "Post",
			expected: true,
			expRes:   "",
			ctrl: ctrl{
				path:   "/post",
				method: "POST",
				body:   "testbody",
			},
			conf: `
url: http://test/post
method: post
body:
  string: testbody
`,
		},
		{
			name:     "PostJSON",
			expected: true,
			expRes:   "",
			ctrl: ctrl{
				path:   "/postjson",
				method: "POST",
				body:   `{"k1":"v1","k2":"v2"}`,
				ct:     "application/json",
			},
			conf: `
url: http://test/postjson
method: post
body:
  json:
    k1: v1
    k2: v2
`,
		},
		{
			name:     "GetJson",
			expected: true,
			ctrl: ctrl{
				path:   "/getjson",
				method: "GET",
				resp:   `{"k1":"v1","k2":"v2"}`,
			},
			expRes: map[string]interface{}{"k1": "v1", "k2": "v2"},
			conf: `
url: http://test/getjson
method: get
response: json
`,
		},
	}
	zerolog.SetGlobalLevel(logLevel)
	log := zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr}).
		With().Caller().Timestamp().Logger()
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			p := New()
			cfg := getConf(t, tc.conf)
			c := ctx.New()
			init := genapid.InitPredicate(log, c, p, cfg)
			assert.Equal(t, !tc.invalidParam, init, "initPredicate")
			if init {
				tc.ctrl.t = t
				srv := HTTPMock(tc.ctrl)
				t.Cleanup(srv.Close)
				assert.Nil(t, setHostURL(p, srv.URL))
				assert.Equal(t,
					tc.expected, p.Call(log, c), "predicate result")
				assert.Equal(t, tc.expRes, p.Result()["response"], "bad response")
				assert.Equal(t, http.StatusOK, p.Result()["code"], "bad code")
			}
		})

	}
}

/***************************************************************************
  HTTP server mock
  ***************************************************************************/

func (c *ctrl) mockHandler(w http.ResponseWriter, r *http.Request) {
	assert.Equal(c.t, c.path, r.URL.Path, "wrong URL path")
	assert.Equal(c.t, c.method, r.Method, "wrong method")
	if c.ct != "" {
		assert.Equal(c.t, c.ct, r.Header.Get("Content-Type"), "wrong content-type")
	}
	for k, v := range c.headers {
		assert.Equal(c.t, v, r.Header.Get(k), "wrong header %v", k)
	}
	for k, v := range c.params {
		assert.Equal(c.t, v, r.URL.Query().Get(k), "wrong URL query %v", k)
	}

	if c.basicAuth != nil {
		user, pass, ok := r.BasicAuth()
		assert.Equal(c.t, true, ok, "missing basic auth")
		assert.Equal(c.t, c.basicAuth, []string{user, pass}, "wrong basic auth")
	}
	body, err := ioutil.ReadAll(r.Body)
	assert.Nil(c.t, err)
	assert.Equal(c.t, c.body, string(body), "wrong body")

	resp := []byte(c.resp)
	_, err = w.Write(resp)
	assert.Nil(c.t, err, "response")
}

func HTTPMock(ctrl ctrl) *httptest.Server {
	handler := http.NewServeMux()
	handler.HandleFunc("/", ctrl.mockHandler)

	return httptest.NewServer(handler)
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
