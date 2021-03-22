package readfilepredicate

import (
	"os"
	"testing"

	"github.com/go-test/deep"
	"github.com/jsautret/go-api-broker/ctx"
	"github.com/jsautret/go-api-broker/genapid"
	"github.com/jsautret/go-api-broker/internal/conf"
	"github.com/kr/pretty"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

var logLevel = zerolog.FatalLevel

func TestMatch(t *testing.T) {
	cases := []struct {
		name         string      // Test name
		conf         string      // YAML conf
		expResult    bool        // predicate result
		invalidParam bool        // Params values are invalid
		expResults   interface{} // regexp matches
	}{
		{
			name:         "NoConf",
			conf:         "",
			invalidParam: true,
		},
		{
			name: "YAMLandJSON",
			conf: `
yaml: testdata/test1.yaml
json: testdata/test1.json
`,
			invalidParam: true,
		},
		{
			name: "NoFileYAML",
			conf: `
yaml: testdata/nofile.yaml
`,
			expResult: false,
		},
		{
			name: "InvalidYAML",
			conf: `
yaml: testdata/invalid
`,
			expResult: false,
		},
		{
			name: "InvalidJSON",
			conf: `
json: testdata/invalid
`,
			expResult: false,
		},
		{
			name: "NoFileJSON",
			conf: `
json: testdata/nofile.yaml
`,
			expResult: false,
		},
		{
			name: "YAMLEmpty",
			conf: `
yaml: testdata/empty
`,
			expResult: true,
		},
		{
			name: "YAML",
			conf: `
yaml: testdata/test1.yaml
`,
			expResult: true,
			expResults: map[string]interface{}{
				"key1": "value1",
				"key2": []interface{}{"value21", "value22"},
				"key3": map[string]interface{}{"key31": "value31", "key32": "value32", "key33": "value33"},
			},
		},
		{
			name: "JSON",
			conf: `
json: testdata/test1.json
`,
			expResult: true,
			expResults: map[string]interface{}{
				"key1": "value1",
				"key2": []interface{}{"value21", "value22"},
				"key3": map[string]interface{}{"key31": "value31", "key32": "value32", "key33": "value33"},
			},
		},
	}

	zerolog.SetGlobalLevel(logLevel)
	log.Logger = zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr}).
		With().Caller().Timestamp().Logger()
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			p := New()
			cfg := getConf(t, tc.conf)
			c := ctx.New()
			init := genapid.InitPredicate(log.Logger, c, p, cfg)
			assert.Equal(t, !tc.invalidParam, init, "initPredicate")
			if init {
				assert.Equal(t,
					tc.expResult, p.Call(log.Logger), "predicate result")
				if diff := deep.Equal(tc.expResults, p.Result()["content"]); diff != nil {
					t.Log(diff)
				}
				assert.Equal(t, tc.expResults,
					p.Result()["content"],
					"mismatches content")
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
