package predicate

import (
	"os"
	"testing"

	"github.com/jsautret/go-api-broker/ctx"
	"github.com/jsautret/go-api-broker/internal/conf"
	"github.com/jsautret/go-api-broker/internal/conf/mocks"
	"github.com/jsautret/go-api-broker/internal/plugins"
	"github.com/kr/pretty"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

var logLevel = zerolog.ErrorLevel

//var logLevel = zerolog.TraceLevel

type testData struct {
	name         string
	conf         string
	expPredicate string
	expResult    bool
	expVars      ctx.Variables
	expConf      map[string]interface{}
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
test1:
  key: value
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
			name:         "set",
			expPredicate: "test_set",
			expResult:    true,
			expConf: map[string]interface{}{
				"key": "value",
			},
			expVars: ctx.Variables{
				"v1": "val1",
				"v2": "val2",
			},
			conf: `
test_set:
  key: value
set:
  - v1: val1
  - v2: val2
`,
		},
	}
	for _, tc := range tt {
		var plugin mocks.Plugin

		// Init context & create conf
		ctx := ctx.New()
		cfg := getConf(t, tc.conf)

		// Register plugin
		plugins.Add(tc.expPredicate, &plugin)
		// Expectations on predicate Call
		expConf := conf.Params{Name: tc.expPredicate, Conf: tc.expConf}
		plugin.On("Call", ctx, &expConf).Return(true)

		// Evaluate
		assert.Equal(t,
			Process(cfg, ctx), tc.expResult,
			"Wrong return for test "+tc.name)

		if tc.expVars != nil {
			assert.Equal(t, tc.expVars, ctx.V, "Variables mismatch")
		}

		// Check Mock
		plugin.AssertExpectations(t)
	}
}

func TestMain(m *testing.M) {
	zerolog.SetGlobalLevel(logLevel)
	log.Logger = zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr}).
		With().Caller().Timestamp().Logger()
	os.Exit(m.Run())
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
