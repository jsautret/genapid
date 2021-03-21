package matchpredicate

import (
	"os"
	"testing"

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
		name         string            // Test name
		conf         string            // YAML input
		expResult    bool              // predicate result
		invalidParam bool              // Params values are invalid
		expResults   []string          // regexp matches
		expNamed     map[string]string // regexp named groups
	}{
		{
			name:         "NoConf",
			conf:         "",
			invalidParam: true,
		},
		{
			name:         "NoString",
			invalidParam: true,
			conf: `
fixed: ABCD
`,
		},
		{
			name: "OnlyString",
			conf: `
string: ABCD
`,
			invalidParam: true,
		},
		{
			name: "FixedMatched",
			conf: `
string: "ABCD"
fixed: ABCD
`,
			expResult: true,
		},
		{
			name: "FixedNotMatched",
			conf: `
string: "ABCDE"
fixed: ABCD
`,
			expResult: false,
		},
		{
			name: "EmptyFixed",
			conf: `
string: ""
fixed: ""
`,
			invalidParam: true,
		},
		{
			name: "FixedAndRegexp",
			conf: `
string: "AAA"
fixed: "AAA"
regexp: "XXX"
`,
			expResult: true,
		},
		{
			name: "EmptyRegexp",
			conf: `
string: ""
regexp: ""
`,
			invalidParam: true,
		},
		{
			name: "BadRegexp",
			conf: `
string: "AAAAA"
regexp: "AA(AA"
`,
			expResult: false,
		},
		{
			name: "RegexpMatched",
			conf: `
string: ABBBBCD
regexp: A(B+.)D$
`,
			expResult:  true,
			expResults: []string{"ABBBBCD", "BBBBC"},
		},
		{
			name: "RegexpNotMatched",
			conf: `
string: ABBBBCDE
regexp: A(B+.)D$
`,
			expResult: false,
		},
		{
			name: "Named",
			conf: `
string: RRRRTTTTSSYYYY
regexp: ^(?P<r>R+)(T+)S*(?P<y>Y+)$
`,
			expResult: true,
			expResults: []string{"RRRRTTTTSSYYYY",
				"RRRR", "TTTT", "YYYY"},
			expNamed: map[string]string{"r": "RRRR", "y": "YYYY"},
		},

		{
			name: "GVal",
			conf: `
string: '= ( 42 < 8 ? "AAAA" : "WWWW") + "AA"'
fixed:  "WWWWAA"
`,
			expResult: true,
		},
		{
			name: "jsonpath",
			conf: `
string: '= {"name": "value"}| $.name'
fixed:  "value"
`,
			expResult: true,
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
					tc.expResult, p.Call(log.Logger))
				if len(tc.expResults) > 0 {
					assert.Equal(t, tc.expResults,
						p.Result()["matches"], "mismatched groups")
					if tc.expNamed != nil {
						assert.Equal(t, tc.expNamed,
							p.Result()["named"],
							"mismatches named group")
					}
				}
			}
		})

	}
}

/***************************************************************************
  Benchmarck: compare predicates with and without templating
  ***************************************************************************/

func BenchmarkNoGval(b *testing.B) {
	yaml := `
string: AAAAAA
fixed:  AAAAAA
`
	benchmark(b, yaml)
}

func BenchmarkWithGval(b *testing.B) {
	yaml := `
string: '= ( 42 < 8 ? "AAAA" : "WWWW") + "AA"'
fixed:  "WWWWAA"
`
	benchmark(b, yaml)
}

func benchmark(b *testing.B, yaml string) {
	p := New()
	cfg := getConfB(b, yaml)
	c := ctx.New()
	zerolog.SetGlobalLevel(logLevel)
	for i := 0; i < b.N; i++ {
		require.True(b,
			genapid.InitPredicate(log.Logger, c, p, cfg))
		p.Call(log.Logger)

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

func getConfB(b *testing.B, source string) *conf.Params {
	c := conf.Params{}
	require.Nil(b,
		yaml.Unmarshal([]byte(source), &c.Conf), "YAML parsing failed")
	return &c
}
