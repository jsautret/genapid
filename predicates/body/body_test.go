// Copyright 2021 Jérôme Sautret. All rights reserved.  Use of this
// source code is governed by an Apache License 2.0 that can be found
// in the LICENSE file.

package bodypredicate

import (
	"io/ioutil"
	"net/http"
	"os"
	"strings"
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

func TestBody(t *testing.T) {
	cases := []struct {
		name         string
		conf         string
		expected     bool        // return of predicate
		invalidParam bool        // true if params values are invalid
		expRes       interface{} // expected predicate result
		method       string      // set in request
		body         string      // body to set in request
		ct           string      // content-type to set in request
	}{
		{
			name:     "NoConf",
			conf:     "",
			method:   "POST",
			expected: true,
		},
		{
			name:         "badType",
			invalidParam: true,
			method:       "POST",
			conf: `
type: badType
`,
		},
		{
			name:     "contentType",
			method:   "POST",
			ct:       "application/json; charset=utf-8",
			expected: true,
			conf: `
mime: application/json
`,
		},
		{
			name:     "badContentType",
			method:   "POST",
			ct:       "application/json; charset=utf-8",
			expected: false,
			conf: `
mime: application/xml
`,
		},
		{
			name:     "bodyString",
			method:   "POST",
			expected: true,
			body:     "value1",
			expRes:   "value1",
			conf: `
type: string
`,
		},
		{
			name:     "bodyStringLimit",
			method:   "POST",
			expected: true,
			body:     "12345",
			expRes:   "12345",
			conf: `
type: string
limit: 5
`,
		},
		{
			name:     "bodyStringLimitReached",
			method:   "POST",
			expected: true,
			body:     "12345",
			expRes:   "1234",
			conf: `
type: string
limit: 4
`,
		},
		{
			name:     "bodyJSON",
			method:   "POST",
			expected: true,
			body:     `{"k1":"v1", "k2":"v2"}`,
			expRes:   map[string]interface{}{"k1": "v1", "k2": "v2"},
			conf: `
type: json
`,
		},
		{
			name:     "bodyJSONBadMethod",
			method:   "GET",
			expected: false,
			body:     `{"k1":"v1", "k2":"v2"}`,
			conf: `
type: json
`,
		},
		{
			name:     "bodyJSONNoType",
			method:   "POST",
			expected: false,
			body:     `{"k1":"v1", "k2":"v2"}`,
			conf: `
type: json
mime: application/json
`,
		},
		{
			name:     "bodyJSONGoodType",
			method:   "POST",
			expected: true,
			body:     `{"k1":"v1", "k2":"v2"}`,
			ct:       "application/json",
			expRes:   map[string]interface{}{"k1": "v1", "k2": "v2"},
			conf: `
type: json
mime: application/json
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
			Req := http.Request{Method: tc.method}
			if tc.ct != "" {
				h := http.Header{"Content-Type": []string{tc.ct}}
				Req.Header = h
			}
			Req.Body = ioutil.NopCloser(strings.NewReader(tc.body))
			c.In = &Req
			init := genapid.InitPredicate(log, c, p, cfg)
			assert.Equal(t, !tc.invalidParam, init, "initPredicate")
			if init {
				assert.Equal(t,
					tc.expected, p.Call(log, c), "bad predicate result")
				assert.Equal(t, tc.expRes, p.Result()["payload"], "bad payload")
			}
		})

	}
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
