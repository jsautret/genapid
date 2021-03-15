package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/jsautret/go-api-broker/internal/conf"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"
)

var logLevel = zerolog.FatalLevel

//var logLevel = zerolog.TraceLevel

func TestHttpServer(t *testing.T) {
	tt := []struct {
		name       string
		method     string
		path       string
		want       string
		statusCode int
		conf       string
	}{
		{
			name:       "EmptyConf",
			method:     http.MethodGet,
			path:       "/EmptyConf",
			statusCode: http.StatusNotFound,
			conf:       "",
		},
		{
			name:       "PipeOfMatch",
			method:     http.MethodGet,
			path:       "/PipeOfMatch",
			statusCode: http.StatusOK,
			conf: `
- name: "Test simple pipe of match"
  pipe:
  - match:
      string:  AAAAAA
      fixed:  "AAAAAA"
  - match:
      string:  ="CCCCC"
      fixed:  CCCCC
  - match:
      string: '=(-1<0 ? "BBBAAABBB" : "BBBBB")'
      regexp: "B(A+B)B"
    register: some_test
  - match:
      string: =R.some_test.matches[1]
      fixed: AAAB
`,
		},
		{
			name:       "IncomingHttpMatching",
			method:     http.MethodPost,
			path:       "/test?param1=value1&param2=value2",
			statusCode: http.StatusOK,
			conf: `
- name: "Test pipe of match on incoming request"
  pipe:
  - match:
      string: =Req.Method
      fixed: POST
  - match:
      string: =Url.Params.param2[0]
      fixed: value2
  - match:
      string: =Url.Params.param1[0]
      regexp: value[1-9]
`,
		},
	}
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			config = getConf(t, tc.conf)
			request := httptest.NewRequest(tc.method, tc.path, nil)
			responseRecorder := httptest.NewRecorder()

			handler(responseRecorder, request)

			if responseRecorder.Code != tc.statusCode {
				t.Errorf("Want status '%d', got '%d'",
					tc.statusCode, responseRecorder.Code)
			}

			if strings.TrimSpace(
				responseRecorder.Body.String()) != tc.want {
				t.Errorf("Want '%s', got '%s'",
					tc.want, responseRecorder.Body)
			}
		})
	}
}

// TODO: test log content with https://github.com/rzajac/zltest

func TestMain(m *testing.M) {
	zerolog.SetGlobalLevel(logLevel)
	log.Logger = zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr}).
		With().Caller().Timestamp().Logger()
	//log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	os.Exit(m.Run())
}

/***************************************************************************
  Benchmarck: compare predicates with and without gval
  ***************************************************************************/
func BenchmarkNoTemplate(b *testing.B) {
	conf := `
- name: "Test simple pipe of match without expressions"
  pipe:
  - match:
      string: "AAAAAA"
      fixed:  "AAAAAA"
  - match:
      string: "BBBAAABBB"
      regexp: "B(A+B)B"
    register: some_test
  - match:
      string: "AAAB"
      fixed: AAAB
`
	zerolog.SetGlobalLevel(logLevel)
	config = getConfB(conf)
	request := httptest.NewRequest(http.MethodGet, "/bench", nil)
	responseRecorder := httptest.NewRecorder()
	for i := 0; i < b.N; i++ {
		handler(responseRecorder, request)
	}

}
func BenchmarkWithTemplate(b *testing.B) {
	conf := `
- name: "Test simple pipe of match with expressions"
  pipe:
  - match:
      string: ="AAAAAA"
      fixed:  "AAAAAA"
  - match:
      string: '= (-2 < -1 ? "BBBAAABBB" : "BBBBB" )'
      regexp: "B(A+B)B"
    register: some_test
  - match:
      string: '= R.some_test.matches[1]'
      fixed: AAAB
`
	zerolog.SetGlobalLevel(zerolog.FatalLevel)
	config = getConfB(conf)
	request := httptest.NewRequest(http.MethodGet, "/bench", nil)
	responseRecorder := httptest.NewRecorder()
	for i := 0; i < b.N; i++ {
		handler(responseRecorder, request)
	}

}

/***************************************************************************
  Helpers
  ***************************************************************************/
func getConf(t *testing.T, source string) conf.Root {
	c := conf.Root{}
	//log.Debug().Str("source", source).Msg("XXX")
	if err := yaml.Unmarshal([]byte(source), &c); err != nil {
		t.Fatalf("YAML parsing: %v", err)
	}
	//log.Debug().Interface("yaml", c).Msg("XXX")
	return c
}

func getConfB(source string) conf.Root {
	c := conf.Root{}
	yaml.Unmarshal([]byte(source), &c)
	return c
}
