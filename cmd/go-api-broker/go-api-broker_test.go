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

func TestHttpServer(t *testing.T) {
	tt := []struct {
		name       string
		method     string
		want       string
		path       string
		statusCode int
		conf       string
	}{
		{
			name:       "Empty conf",
			method:     http.MethodGet,
			path:       "/",
			statusCode: http.StatusNotFound,
			conf:       "",
		},
		{
			name:       "Pipe of match",
			method:     http.MethodGet,
			path:       "/",
			statusCode: http.StatusOK,
			conf: `
- name: "Test simple pipe of match"
  pipe:
  - match:
      string: "AAAAAA"
      fixed:  "AAAAAA"
  - match:
      string: "BBBAAABBB"
      regexp: "B(A+B)B"
    register: some_test
  - match:
      string: "{{index .R.some_test.matches 1}}"
      fixed: AAAB
`,
		},
		{
			name:       "Incoming HTTP matching",
			method:     http.MethodPost,
			path:       "/test?param1=value1&param2=value2",
			statusCode: http.StatusOK,
			conf: `
- name: "Test pipe of match on incoming request"
  pipe:
  - match:
      string: "{{.Req.Method}}"
      fixed: POST
  - match:
      string: "{{.Url.Params.param2}}"
      regexp: value2
  - match:
      string: "{{.Url.Params.param1}}"
      regexp: value1
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
	zerolog.SetGlobalLevel(zerolog.FatalLevel)
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	os.Exit(m.Run())
}

//**************** Helpers
func getConf(t *testing.T, source string) conf.Root {
	c := conf.Root{}
	if err := yaml.Unmarshal([]byte(source), &c); err != nil {
		t.Errorf("Should not have returned parsing error")
	}
	return c
}
