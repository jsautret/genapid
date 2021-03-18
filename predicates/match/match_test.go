package matchpredicate

import (
	"os"
	"testing"

	"github.com/jsautret/go-api-broker/ctx"
	"github.com/jsautret/go-api-broker/internal/conf"
	"github.com/kr/pretty"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

var logLevel = zerolog.FatalLevel

//var logLevel = zerolog.DebugLevel

func TestMatch(t *testing.T) {
	cases := []struct {
		name     string
		conf     string
		expected bool
	}{
		{
			name:     "NoConf",
			conf:     "",
			expected: false,
		},
		{
			name: "NoString",
			conf: `
fixed: ABCD
`,
			expected: false,
		},
		{
			name: "OnlyString",
			conf: `
string: ABCD
`,
			expected: false,
		},
		{
			name: "FixedMatched",
			conf: `
string: "ABCD"
fixed: ABCD
`,
			expected: true,
		},
		{
			name: "FixedNotMatched",
			conf: `
string: "ABCDE"
fixed: ABCD
`,
			expected: false,
		},
		{
			name: "EmptyFixed",
			conf: `
string: ""
fixed: ""
`,
			expected: false,
		},
		{
			name: "FixedAndRegexp",
			conf: `
string: "AAA"
fixed: "AAA"
regexp: "XXX"
`,
			expected: true,
		},
		{
			name: "EmptyRegexp",
			conf: `
string: ""
regexp: ""
`,
			expected: false,
		},
		{
			name: "BadRegexp",
			conf: `
string: "AAAAA"
regexp: "AA(AA"
`,
			expected: false,
		},
		{
			name: "RegexpMatched",
			conf: `
string: ABBBBCD
regexp: A(B+.)D$
`,
			expected: true,
		},
		{
			name: "RegexpNotMatched",
			conf: `
string: ABBBBCDE
regexp: A(B+.)D$
`,
			expected: false,
		},
		{
			name: "Templating",
			conf: `
string: '= ( 42 < 8 ? "AAAA" : "WWWW") + "AA"'
fixed:  "WWWWAA"
`,
			expected: true,
		},
		{
			name: "jsonpath",
			conf: `
string: '= {"name": "value"}| $.name'
fixed:  "value"
`,
			expected: true,
		},
	}

	for _, c := range cases {
		var self Predicate
		t.Run(c.name, func(t *testing.T) {
			conf := getConf(t, c.conf)
			ctx := ctx.New()
			if r := self.Call(ctx, conf); r != c.expected {
				t.Errorf("Should have returned %v, got %v",
					c.expected, r)
			}
		})

	}
}

func TestMatchRegexpWithRegister(t *testing.T) {
	var self Predicate

	yaml := `
string: ABBBBCD
regexp: A(B+.)D$
xxx: ccc
`
	conf := getConf(t, yaml)
	ctx := ctx.New()
	if res := self.Call(ctx, conf); !res {
		t.Errorf("Should have returned true")
	} else {
		if r := self.Result()["matches"].([]string); len(r) == 0 ||
			r[1] != "BBBBC" {
			t.Errorf("Should have match BBBBC, not %v", r)
		}
	}

}

func TestMain(m *testing.M) {
	zerolog.SetGlobalLevel(logLevel)
	log.Logger = zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr}).
		With().Caller().Timestamp().Logger()
	os.Exit(m.Run())
}

/***************************************************************************
  Benchmarck: compare predicates with and without templating
  ***************************************************************************/
func BenchmarkNoGval(b *testing.B) {
	var self Predicate
	yaml := `
string: AAAAAA
fixed:  AAAAAA
`
	zerolog.SetGlobalLevel(logLevel)
	conf := getConfB(b, yaml)
	ctx := ctx.New()
	for i := 0; i < b.N; i++ {
		self.Call(ctx, conf)
	}
}
func BenchmarkWithGval(b *testing.B) {
	var self Predicate
	yaml := `
string: '= ( 42 < 8 ? "AAAA" : "WWWW") + "AA"'
fixed:  "WWWWAA"
`
	zerolog.SetGlobalLevel(logLevel)
	conf := getConfB(b, yaml)
	ctx := ctx.New()
	for i := 0; i < b.N; i++ {
		self.Call(ctx, conf)
	}

}

/***************************************************************************
  Helpers
  ***************************************************************************/
func getConf(t *testing.T, source string) *conf.Params {
	c := conf.Params{Name: "test"}
	require.Nil(t,
		yaml.Unmarshal([]byte(source), &c.Conf), "YAML parsing failed")
	t.Logf("Parsed YAML:\n%# v", pretty.Formatter(c))

	return &c
}

func getConfB(b *testing.B, source string) *conf.Params {
	c := conf.Params{Name: "bench"}
	require.Nil(b,
		yaml.Unmarshal([]byte(source), &c.Conf))
	return &c
}
