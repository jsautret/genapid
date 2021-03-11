package match

import (
	"os"
	"testing"

	"github.com/jsautret/go-api-broker/context"
	"github.com/jsautret/go-api-broker/internal/conf"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"
)

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
string: '{{if true}}{{"WWWWWW"| printf "%s"}}{{else}}XXXXXX{{end}}'
fixed:  "WWWWWW"
`,
			expected: true,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			conf := getConf(t, c.conf)
			ctx := context.Ctx{
				R:       make(map[string]map[string]interface{}),
				Results: make(map[string]interface{}),
			}
			if r := Call(&ctx, conf); r != c.expected {
				t.Errorf("Should have returned %v, got %v",
					c.expected, r)
			}
		})

	}
}

func TestMatchRegexpWithRegister(t *testing.T) {
	yaml := `
string: ABBBBCD
regexp: A(B+.)D$
xxx: ccc
`
	conf := getConf(t, yaml)
	ctx := context.Ctx{
		R:       make(map[string]map[string]interface{}),
		Results: make(map[string]interface{}),
	}
	if res := Call(&ctx, conf); !res {
		t.Errorf("Should have returned true")
	} else {
		if ctx.Results["matches"].([]string)[1] != "BBBBC" {
			t.Errorf("Should have match BBBBC, not %v",
				ctx.Results["matches"].([]string)[1])
		}
	}

}

func TestMain(m *testing.M) {
	zerolog.SetGlobalLevel(zerolog.FatalLevel)
	log.Logger = zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr}).
		With().Caller().Timestamp().Logger()
	os.Exit(m.Run())
}

/***************************************************************************
  Benchmarck: compare predicates with and without templating
  ***************************************************************************/
func BenchmarkNoTemplate(b *testing.B) {
	yaml := `
string: AAAAAA
fixed:  AAAAAA
`
	zerolog.SetGlobalLevel(zerolog.FatalLevel)
	conf := getConfB(yaml)
	ctx := context.Ctx{
		R:       make(map[string]map[string]interface{}),
		Results: make(map[string]interface{}),
	}
	for i := 0; i < b.N; i++ {
		Call(&ctx, conf)
	}
}
func BenchmarkWithTemplate(b *testing.B) {
	yaml := `
string: '{{if true}}{{"AAAAAA"| printf "%s"}}{{else}}BBBBB{{end}}'
fixed:  "AAAAAA"
`
	zerolog.SetGlobalLevel(zerolog.FatalLevel)
	conf := getConfB(yaml)
	ctx := context.Ctx{
		R:       make(map[string]map[string]interface{}),
		Results: make(map[string]interface{}),
	}
	for i := 0; i < b.N; i++ {
		Call(&ctx, conf)
	}

}

/***************************************************************************
  Helpers
  ***************************************************************************/
func getConf(t *testing.T, source string) conf.Predicate {
	c := conf.Predicate{}
	if err := yaml.Unmarshal([]byte(source), &c); err != nil {
		t.Errorf("Should not have returned parsing error")
	}
	return c
}

func getConfB(source string) conf.Predicate {
	c := conf.Predicate{}
	yaml.Unmarshal([]byte(source), &c)
	return c
}
