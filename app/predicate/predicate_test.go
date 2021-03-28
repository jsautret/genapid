// Copyright 2021 Jérôme Sautret. All rights reserved.  Use of this
// source code is governed by an Apache License 2.0 that can be found
// in the LICENSE file.

package predicate

import (
	"os"
	"testing"

	"github.com/jsautret/genapid/app/conf"
	"github.com/jsautret/genapid/app/plugins"
	"github.com/jsautret/genapid/ctx"
	"github.com/jsautret/genapid/genapid"
	"github.com/jsautret/genapid/genapid/mocks"
	"github.com/kr/pretty"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

var logLevel = zerolog.TraceLevel

type testData struct {
	name            string
	conf            string
	expPredicate    string
	expResult       bool
	expVars         ctx.Variables
	expDef          ctx.Default
	expConf         map[string]interface{}
	expRegister     string
	expRegisterData ctx.Result
}

func TestPredicate(t *testing.T) {

	tt := []testData{
		{
			name:         "keyValue",
			expPredicate: "test1",
			expResult:    true,
			expConf: map[string]interface{}{
				"key": "value",
			},
			conf: `
name: keyValue
test1:
  key: value
`,
		},
		{
			name:         "wrongPredicate",
			expPredicate: "test1",
			expResult:    false,
			conf: `
name: keyValue
test2:
  key: value
`,
		},
		{
			name:         "wrongParameters",
			expPredicate: "test1",
			expResult:    false,
			conf: `
name: keyValue
test2:
  - key: value
`,
		},
		{
			name:         "wrongResult",
			expPredicate: "test1",
			expResult:    false,
			conf: `
name: keyValue
test1:
  key: value
result: wrong
`,
		},
		{
			name:         "complexValue",
			expPredicate: "test2",
			expResult:    true,
			expConf: map[string]interface{}{
				"key1": []interface{}{"value11", "value12"},
				"key2": map[string]interface{}{
					"key21": "value21",
					"key22": "value22",
				},
			},
			conf: `
test2:
  key1:
    - value11
    - value12
  key2:
    key21: value21
    key22: value22
`,
		},
		{
			name:         "variable",
			expPredicate: "test_variable",
			expResult:    true,
			expVars: ctx.Variables{
				"v1": "val1",
				"v2": "val2",
			},
			conf: `
variable:
  - v1: val1
  - v2: val2
`,
		},
		{
			name:         "variableWhenTrue",
			expPredicate: "test_variable",
			expResult:    true,
			expVars: ctx.Variables{
				"v1": "val1",
				"v2": "val2",
			},
			conf: `
variable:
  - v1: val1
  - v2: val2
when: =true
`,
		},
		{
			name:         "variableWhenFalse",
			expPredicate: "test_variable",
			expResult:    true,
			expVars:      ctx.Variables{},
			conf: `
variable:
  - v1: val1
  - v2: val2
when: =false
`,
		},
		{
			name:         "variableWhenWrong",
			expPredicate: "test_variable",
			expResult:    true,
			expVars:      ctx.Variables{},
			conf: `
variable:
  - v1: val1
  - v2: val2
when: wrong
`,
		},
		{
			name:         "variableAndPredicate",
			expPredicate: "test_variable",
			expResult:    false,
			conf: `
test_variable:
  key: value
variable:
  - v1: val1
  - v2: val2
`,
		},
		{
			name:         "variableAndDefault",
			expPredicate: "test_variable",
			expResult:    false,
			conf: `
default:
  key: value
variable:
  - v1: val1
  - v2: val2
`,
		},
		{
			name:         "defaultAndVariable",
			expPredicate: "test_variable",
			expResult:    false,
			conf: `
variable:
  - v1: val1
  - v2: val2
default:
  key: value
`,
		},
		{
			name:         "default",
			expPredicate: "test_variable",
			expDef:       ctx.Default{"test_default": ctx.DefaultParams{"key": "value"}},
			expResult:    true,
			conf: `
default:
  test_default:
    key: value
`,
		},
		{
			name:         "defaultAndPredicate",
			expPredicate: "test_default",
			expResult:    false,
			conf: `
test_default:
  key: value
default:
  test_default:
    key: value
`,
		},
		{
			name:         "register",
			expPredicate: "test_register",
			expResult:    true,
			expConf: map[string]interface{}{
				"key": "value",
			},
			expRegister:     "registered",
			expRegisterData: ctx.Result{"k": "v"},
			conf: `
test_register:
  key: value
register: registered
`,
		},
		{
			name:         "registerWrong",
			expPredicate: "test_register",
			expResult:    false,
			expRegister:  "registered",
			conf: `
test_register:
  key: value
register:
  - s1
  - s2
`,
		},
		{
			name:         "registerEmpty",
			expPredicate: "test_register_empty",
			expResult:    true,
			expConf: map[string]interface{}{
				"key": "value",
			},
			expRegister: "empty",
			conf: `
test_register_empty:
  key: value
register: empty
`,
		},
		{
			name:         "pipe",
			expPredicate: "test_variable",
			expResult:    true,
			expVars: ctx.Variables{
				"v1": "val1",
				"v2": "val2",
			},
			conf: `
name: pipe
pipe:
  - variable:
     - v1: val1
     - v2: val2
`,
		},
		{
			name:         "pipeWrong",
			expPredicate: "test_variable",
			expResult:    false,
			conf: `
name: pipe
pipe:
  variable:
   - v1: val1
   - v2: val2
`,
		},
	}

	zerolog.SetGlobalLevel(logLevel)
	log.Logger = zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr}).
		With().Caller().Timestamp().Logger()
	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {

			p := mocks.Predicate{}

			// Init context & create conf
			c := ctx.New()
			cfg := getConf(t, tc.conf)

			p.On("Name").Return(tc.expPredicate)
			// Register plugin
			plugins.Add(p.Name(), func() genapid.Predicate { return &p })

			p.On("Call", mock.Anything, mock.Anything).Return(true)
			parsedConf := map[string]interface{}{}
			p.On("Params").Return(&parsedConf)
			if tc.expRegister != "" {
				p.On("Result").Return(tc.expRegisterData)
			}

			// Evaluate
			res := Process(log.Logger, cfg, c)
			// Check boolean result
			assert.Equal(t, tc.expResult, res,
				"Wrong return for Process")
			if res {
				switch {
				case tc.expVars != nil:
					// Check variables set by predicate
					assert.Equal(t, tc.expVars, c.V,
						"Variables mismatch")
				case tc.expDef != nil:
					// Check default set by predicate
					assert.Equal(t, tc.expDef, c.Default,
						"Default mismatch")
				default:
					// Check params received by predicate
					assert.Equal(t, tc.expConf, parsedConf)
					if tc.expRegister != "" {
						// Check registered data
						exp := tc.expRegisterData
						if exp != nil {
							// We add boolean result
							exp["result"] = res
						} else {
							// No data is expected, we
							// just expect boolean result
							exp = ctx.Result{"result": res}
						}
						assert.Equal(t,
							ctx.Registered{tc.expRegister: exp}, c.R)
					}
					// Check Mock calls
					p.AssertExpectations(t)
				}
			}
		})
	}
}

/***************************************************************************
  Helpers
  ***************************************************************************/
func getConf(t *testing.T, source string) *conf.Predicate {
	c := conf.Predicate{}
	require.Nil(t,
		yaml.Unmarshal([]byte(source), &c))
	t.Logf("Parsed YAML:\n%# v", pretty.Formatter(c))
	return &c
}
