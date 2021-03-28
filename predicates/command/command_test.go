// Copyright 2021 Jérôme Sautret. All rights reserved.  Use of this
// source code is governed by an Apache License 2.0 that can be found
// in the LICENSE file.

// +build !windows

package commandpredicate

import (
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

func TestHTTP(t *testing.T) {
	cases := []struct {
		name         string
		conf         string
		expected     bool // return of predicate
		invalidParam bool // true if params values are invalid
		expRc        int  // expected command rc
		expStdout    string
		expStderr    string
	}{
		{
			name:         "NoConf",
			conf:         "",
			invalidParam: true,
		},
		{
			name: "true",
			conf: `
command: "true"
`,
			expected: true,
		},
		{
			name: "true",
			conf: `
command: "false"
`,
			expected: false,
			expRc:    1,
		},
		{
			name:      "tr",
			expRc:     0,
			expStdout: "hello world!",
			expStderr: "",
			conf: `
command: tr
args:
  - A-Z
  - a-z
stdin: "Hello World!"
`,
			expected: true,
		},
		{
			name:      "pwd",
			expRc:     0,
			expStdout: "/usr\n",
			expStderr: "",
			conf: `
command: pwd
chdir: /usr
`,
			expected: true,
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
				assert.Equal(t,
					tc.expected, p.Call(log, c), "bad predicate result")
				assert.Equal(t, tc.expRc, p.Result()["rc"], "bad rc")
				assert.Equal(t, tc.expStdout, p.Result()["stdout"], "bad stdout")
				assert.Equal(t, tc.expStderr, p.Result()["stderr"], "bad stderr")
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
