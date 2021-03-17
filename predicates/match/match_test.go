package matchpredicate

import (
	"os"
	"testing"

	"github.com/jsautret/go-api-broker/ctx"
	"github.com/jsautret/go-api-broker/internal/conf"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
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
			ctx := ctx.Ctx{
				R:       make(map[string]map[string]interface{}),
				Results: make(map[string]interface{}),
			}
			if r := self.Call(&ctx, conf); r != c.expected {
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
	ctx := ctx.Ctx{
		R:       make(map[string]map[string]interface{}),
		Results: make(map[string]interface{}),
	}
	if res := self.Call(&ctx, conf); !res {
		t.Errorf("Should have returned true")
	} else {
		if ctx.Results["matches"].([]string)[1] != "BBBBC" {
			t.Errorf("Should have match BBBBC, not %v",
				ctx.Results["matches"].([]string)[1])
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
	conf := getConfB(yaml)
	ctx := ctx.Ctx{
		R:       make(map[string]map[string]interface{}),
		Results: make(map[string]interface{}),
	}
	for i := 0; i < b.N; i++ {
		self.Call(&ctx, conf)
	}
}
func BenchmarkWithGval(b *testing.B) {
	var self Predicate
	yaml := `
string: '= ( 42 < 8 ? "AAAA" : "WWWW") + "AA"'
fixed:  "WWWWAA"
`
	zerolog.SetGlobalLevel(logLevel)
	conf := getConfB(yaml)
	ctx := ctx.Ctx{
		R:       make(map[string]map[string]interface{}),
		Results: make(map[string]interface{}),
	}
	for i := 0; i < b.N; i++ {
		self.Call(&ctx, conf)
	}

}

/***************************************************************************
  Helpers
  ***************************************************************************/
func getConf(t *testing.T, source string) conf.Params {
	c := conf.Params{Name: "test"}
	if err := yaml.Unmarshal([]byte(source), &c.Conf); err != nil {
		t.Errorf("Should not have returned parsing error")
	}
	return c
}

func getConfB(source string) conf.Params {
	c := conf.Params{}
	yaml.Unmarshal([]byte(source), &c)
	return c
}
