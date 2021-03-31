// Copyright 2021 Jérôme Sautret. All rights reserved.  Use of this
// source code is governed by an Apache License 2.0 that can be found
// in the LICENSE file.

package main

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/jsautret/genapid/app/conf"
	"github.com/jsautret/genapid/ctx"
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
			statusCode: http.StatusOK,
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
      value:  "AAAAAA"
  - match:
      string:  ="CCCCC"
      value:  CCCCC
  - match:
      string: '=(-1<0 ? "BBBAAABBB" : "BBBBB")'
      regexp: "B(A+B)B"
    register: some_test
  - match:
      string: =R.some_test.matches[1]
      value: AAAB
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
      string: =In.Method
      value: POST
  - match:
      string: '=In.URL.Query()|param2[0]'
      value: value2
  - match:
      string: '=In.URL.Query()|param1[0]'
      regexp: value[1-9]
  - body:
      mime: application/json
      type: json
    register: body
  - match:
      string: =R.body.payload.bodyName
      value: bodyValue
  - body:
      mime: application/json
      type: json
    register: body2
  - match:
      string: =R.body2.payload.bodyName
      value: bodyValue
  - log:
      msg: EndIncomingHttpMatching
`,
		},
		{
			name:       "DefaultFields",
			method:     http.MethodGet,
			path:       "/PipeOfMatch",
			statusCode: http.StatusOK,
			logFound: expLog{
				{"log": "SubPipe"},
				{"log": "EndDefaultFields"},
			},
			conf: `
- name: "Test DefaultFields"
  pipe:
  - default:
      match:
        string: DDDDDD
  - match:
      value:  "DDDDDD"
  - match:
      string:  CCCCC
      value:  CCCCC
  - pipe:
    - default:
        match:
          string: FFFFF
    - match:
        value: FFFFF
    - log:
        msg: SubPipe
  - match:
      regexp: D+
  - default:
      match:
        string: EEEEEE
  - match:
      value: EEEEEE
  - default:
      match:
        value: EEEEEE
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
        value:  imbricated
    - match:
        string: A
        value: B
  # top-level pipe will continue
  # even if predicate fails in sub-pipe
  - match:
      value: imbricated
  - log:
      msg: EndImbricatedPipes
`,
		},
		{
			name:       "InvalidPredicate",
			method:     http.MethodGet,
			path:       "/InvalidPredicate",
			statusCode: http.StatusOK,
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
			statusCode:  http.StatusOK,
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
        value: stop
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
		{
			name:       "Init",
			method:     http.MethodGet,
			path:       "/Init",
			statusCode: http.StatusOK,
			logFound: expLog{
				{"log": "Start"},
				{"log": "End"},
			},
			conf: `
- init:
    - log:
        msg: Start
    - variable:
        - name1: value1
        - name2: value2
    - match:
        string: "BBBAAABBB"
        regexp: "B(A+)B"
      register: match
    - default:
        match:
          value: value1

- name: "Init"
  pipe:
    - match:
       string: "=V.name1"
    - match:
       string: "=R.match.matches[1]"
       value: AAA
    - log:
        msg: End
`,
		},
		{
			name:       "InitWithPipe",
			method:     http.MethodGet,
			path:       "/InitWithPipe",
			statusCode: http.StatusOK,
			logFound: expLog{
				{"log": "Start"},
				{"error": "'pipe' cannot be used in 'init' section"},
			},
			logNotFound: expLog{
				{"log": "Pipe"},
				{"log": "Init"},
				{"log": "End"},
			},
			conf: `
- pipe: # is illegal here because an init is defined
    - log: # won't be evaluated
        msg: Pipe
  init: # won't be evaluated because of illegal pipe
    - log:
        msg: Init
    - variable:
        - name1: value1
        - name2: value2
- name: "InitWithPipe"
  pipe:
    - log:
        msg: Start
    - match:
       string: "=V.name1" # error, init was not evaluated
       value: value1
    - log:
        msg: End
`,
		},
		{
			name:       "InitInPipe",
			method:     http.MethodGet,
			path:       "/InitInPipe",
			statusCode: http.StatusOK,
			logFound: expLog{
				{"log": "Pipe1"},
				{"error": "Cannot use 'init' with a 'pipe'"},
				{"log": "Pipe2"},
			},
			logNotFound: expLog{
				{"log": "Init"},
				{"log": "End"},
			},
			conf: `
- name: "Pipe1"
  pipe:
    - log:
        msg: Pipe1
- init: # is illegal if not at the start, it won't be evaluated
    - log:
        msg: Init
    - variable:
        - name1: value1
        - name2: value2
  name: Pipe2
  pipe:
    - log:
        msg: Pipe2
    - match:
       string: "=V.name1" # error, init was not evaluated
       value: value1
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
			staticCtx = ctx.New()
			processInit(&config, staticCtx)
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
      value:  "AAAAAA"
  - match:
      string: "BBBAAABBB"
      regexp: "B(A+B)B"
    register: some_test
  - match:
      string: "AAAB"
      value: AAAB
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
      value:  "AAAAAA"
  - match:
      string: '= (-2 < -1 ? "BBBAAABBB" : "BBBBB" )'
      regexp: "B(A+B)B"
    register: some_test
  - match:
      string: '= R.some_test.matches[1]'
      value: AAAB
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
	if err := yaml.Unmarshal([]byte(source), &c); err != nil {
		t.Fatalf("YAML parsing: %v", err)
	}
	return c
}

func getConfB(b *testing.B, source string) conf.Root {
	c := conf.Root{}
	require.Nil(b, yaml.Unmarshal([]byte(source), &c))
	return c
}
