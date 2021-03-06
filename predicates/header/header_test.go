// Copyright 2021 Jérôme Sautret. All rights reserved.  Use of this
// source code is governed by an Apache License 2.0 that can be found
// in the LICENSE file.

package headerpredicate

import (
	"net/http"
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

func TestHeader(t *testing.T) {
	cases := []struct {
		name         string
		conf         string
		expected     bool // return of predicate
		invalidParam bool // true if params values are invalid
		header       []string
	}{
		{
			name:         "NoConf",
			conf:         "",
			invalidParam: true,
		},
		{
			name:     "noHeader",
			expected: false,
			conf: `
name: Key
value: value
`,
		},
		{
			name:     "Header",
			expected: true,
			header:   []string{"Key", "value"},
			conf: `
name: Key
value: value
`,
		},
		{
			name:     "Header",
			expected: true,
			header:   []string{"Key", "value"},
			conf: `
name: key
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
			Req := http.Request{}
			if tc.header != nil {
				h := http.Header{tc.header[0]: []string{tc.header[1]}}
				Req.Header = h
			}
			c.In = &Req
			init := genapid.InitPredicate(log, c, p, cfg)
			assert.Equal(t, !tc.invalidParam, init, "initPredicate")
			if init {
				assert.Equal(t,
					tc.expected, p.Call(log, c),
					"bad predicate result")
				if tc.header != nil {
					assert.Equal(t, tc.header[1],
						p.Result()["value"], "bad value")
				}
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
