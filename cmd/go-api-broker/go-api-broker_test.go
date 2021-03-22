package main

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/jsautret/go-api-broker/internal/conf"
	"github.com/jsautret/zltest"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

var logLevel = zerolog.InfoLevel

var checkLog = true

func TestFullConf(t *testing.T) {
	type expLog []map[string]string
	tt := []struct {
		name        string
		method      string
		path        string
		mime        string
		body        string
		want        string
		statusCode  int
		conf        string
		logFound    expLog
		logNotFound expLog
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
			logFound:   expLog{{"log": "EndPipeOfMatch"}},
			conf: `
- name: "Test PipeOfMatch"
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
  - log:
      msg: EndPipeOfMatch
`,
		},
		{
			name:       "IncomingHttpMatching",
			method:     http.MethodPost,
			path:       "/test?param1=value1&param2=value2",
			statusCode: http.StatusOK,
			mime:       "application/json; charset=utf-8",
			body:       `{"bodyName": "bodyValue"}`,
			logFound:   expLog{{"log": "EndIncomingHttpMatching"}},
			conf: `
- name: "Test IncomingHttpMatching"
  pipe:
  - match:
      string: =In.Req.Method
      fixed: POST
  - match:
      string: =In.URL.Params.param2[0]
      fixed: value2
  - match:
      string: =In.URL.Params.param1[0]
      regexp: value[1-9]
  - match:
      string: =In.Mime
      fixed: application/json
  - match:
      string: =jsonpath("$.bodyName", In.Body)
      fixed: bodyValue
  - log:
      msg: EndIncomingHttpMatching
`,
		},
		{
			name:       "DefaultFields",
			method:     http.MethodGet,
			path:       "/PipeOfMatch",
			statusCode: http.StatusOK,
			logFound:   expLog{{"log": "EndDefaultFields"}},
			conf: `
- name: "Test DefaultFields"
  pipe:
  - default:
      match:
        string: DDDDDD
  - match:
      fixed:  "DDDDDD"
  - match:
      string:  CCCCC
      fixed:  CCCCC
  - match:
      regexp: D+
  - default:
      match:
        string: EEEEEE
  - match:
      fixed: EEEEEE
  - default:
      match:
        fixed: EEEEEE
  - match:
  - log:
      msg: EndDefaultFields
`,
		},
		{
			name:       "ImbricatedPipes",
			method:     http.MethodGet,
			path:       "/ImbricatedPipes",
			statusCode: http.StatusOK,
			logFound:   expLog{{"log": "EndImbricatedPipes"}},
			conf: `
- name: "Test ImbricatedPipes"
  pipe:
  - default:
      match:
        string: imbricated
  - name: subPipe1
    pipe:
    - match:
        fixed:  imbricated
    - match:
        string: A
        fixed: B
  # top-level pipe will continue
  # even if predicate fails in sub-pipe
  - match:
      fixed: imbricated
  - log:
      msg: EndImbricatedPipes
`,
		},
		{
			name:       "InvalidPredicate",
			method:     http.MethodGet,
			path:       "/InvalidPredicate",
			statusCode: http.StatusNotFound,
			logFound: expLog{
				{"log": "start"},
				{"error": "Unknown predicate 'invalid'"},
			},
			logNotFound: expLog{{"log": "NotExecuted"}},
			conf: `
- name: "InvalidPredicate"
  pipe:
  - log:
      msg: start
  - invalid:
      key: value
  - log: # will not be evaluated
      msg: NotExecuted
`,
		},
		{
			name:        "StopPipe",
			method:      http.MethodGet,
			path:        "/Stop",
			statusCode:  http.StatusNotFound,
			logFound:    expLog{{"log": "endPipe"}},
			logNotFound: expLog{{"log": "NotExecuted"}},
			conf: `
- name: "StopPipe"
  pipe:
  - default:
      match:
        string: stop
  - name: subPipe1
    pipe:
    - match:
        string: stop
        fixed: stop
    - log:
        msg: endPipe
    result: =false
  - log: # will not be evaluated
      msg: NotExecuted
`,
		},
		{
			name:       "When",
			method:     http.MethodGet,
			path:       "/When",
			statusCode: http.StatusOK,
			logFound: expLog{
				{"log": "Start"},
				{"log": "End"},
			},
			logNotFound: expLog{{"log": "Skipped"}},
			conf: `
- name: "When"
  pipe:
    - log:
        msg: Start
      when: =true
    - when: =false
      log:
        msg: Skipped
    - log:
        msg: End
`,
		},
	}
	for _, tc := range tt {
		zerolog.SetGlobalLevel(logLevel)
		log.Logger = zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr}).
			With().Caller().Timestamp().Logger()

		t.Run(tc.name, func(t *testing.T) {
			var tst *zltest.Tester
			if checkLog {
				tst = zltest.New(t)
				// Configure zerolog and pass tester as a writer.
				log.Logger = zerolog.New(tst).With().
					Timestamp().Logger()
				zerolog.SetGlobalLevel(zerolog.InfoLevel)

			}
			config = getConf(t, tc.conf)

			request := httptest.NewRequest(tc.method, tc.path,
				bytes.NewBuffer([]byte(tc.body)))
			if tc.mime != "" {
				request.Header.Add("Content-Type", tc.mime)
			}
			responseRecorder := httptest.NewRecorder()

			handler(responseRecorder, request)

			if checkLog {
				for _, l := range tc.logFound {
					for k, v := range l {
						tst.Entries().ExpStr(k, v)
					}
				}
				for _, l := range tc.logNotFound {
					for k, v := range l {
						tst.Entries().NotExpStr(k, v)
					}
				}
			}
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

/***************************************************************************
  Benchmarck: compare predicates with and without gval
  ***************************************************************************/
func BenchmarkNoGval(b *testing.B) {
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
	config = getConfB(b, conf)
	request := httptest.NewRequest(http.MethodGet, "/bench", nil)
	responseRecorder := httptest.NewRecorder()
	for i := 0; i < b.N; i++ {
		handler(responseRecorder, request)
	}

}
func BenchmarkWithGval(b *testing.B) {
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
	config = getConfB(b, conf)
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

func getConfB(b *testing.B, source string) conf.Root {
	c := conf.Root{}
	require.Nil(b, yaml.Unmarshal([]byte(source), &c))
	return c
}
