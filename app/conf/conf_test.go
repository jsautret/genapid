// Copyright 2021 Jérôme Sautret. All rights reserved.  Use of this
// source code is governed by an Apache License 2.0 that can be found
// in the LICENSE file.

package conf

import (
	"os"
	"testing"

	"github.com/go-test/deep"
	"github.com/jsautret/genapid/ctx"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

var logLevel = zerolog.InfoLevel

type params struct {
	S1, S2 string
	I      interface{}
	L1     []string
	L2     []interface{}
	N      int
}

func TestConf(t *testing.T) {
	cases := []struct {
		name     string
		conf     string
		expected params
	}{
		{
			name: "Literals",
			conf: `
s1: string1
s2: string2
i: string3
l1:
  - l1
  - l2
`,
			expected: params{
				S1: "string1",
				S2: "string2",
				I:  "string3",
				L1: []string{"l1", "l2"},
			},
		},
		{
			name: "Lists",
			conf: `
s1: string1
s2: string2
i:
  - i1
  - i2
l1:
  - l1
  - l2
l2:
  - l3
  - l4
`,
			expected: params{
				S1: "string1",
				S2: "string2",
				I:  toInterfaceList([]string{"i1", "i2"}),
				L1: []string{"l1", "l2"},
				L2: toInterfaceList([]string{"l3", "l4"}),
			},
		},
		{
			name: "Variables",
			conf: `
s1: '=V.variable1'
`,
			expected: params{
				S1: "value1",
			},
		},
		{
			name: "Request",
			conf: `
s1: '=In.Req.Method'
`,
			expected: params{
				S1: "POST",
			},
		},
		{
			name: "Variables",
			conf: `
s1: '=V.variable1'
s2: '=V.variable2[1]'
`,
			expected: params{
				S1: "value1",
				S2: "value22",
			},
		},
		{
			name: "Jsonpath",
			conf: `
s1: '={"name": "value"}|$.name'
s2: '=jsonpath("$.f2", V.variable3)'
`,
			expected: params{
				S1: "value",
				S2: "value32",
			},
		},
		{
			name: "JsonParsing",
			conf: `
s1: '=jsonpath("$.name", V.variable4)'
`,
			expected: params{
				S1: "value4",
			},
		},
		{
			name: "Fuzzy",
			conf: `
s1: '=fuzzy("whl", V.fuzzy)'
`,
			expected: params{
				S1: "wheel",
			},
		},
		{
			name: "FuzzyNoMatch",
			conf: `
s1: '=fuzzy("x", V.fuzzy)'
`,
			expected: params{
				S1: "",
			},
		},
		{
			name: "Format",
			conf: `
s1: '=format("%s%v", "Result", 42)'
`,
			expected: params{
				S1: "Result42",
			},
		},
		{
			name: "Len",
			conf: `
n: '=len( [1, 2, "foo"] )'
`,
			expected: params{
				N: 3,
			},
		},
		{
			name: "Upper",
			conf: `
s1: '=upper("aB2c,d")'
`,
			expected: params{
				S1: "AB2C,D",
			},
		},
		{
			name: "hmacSha1",
			conf: `
s1: '=hmacSha1("key", "value")'
`,
			expected: params{
				S1: "57443a4c052350a44638835d64fd66822f813319",
			},
		},
		{
			name: "hmacSha256",
			conf: `
s1: '=hmacSha256("key", "value")'
`,
			expected: params{
				S1: "90fbfcf15e74a36b89dbdb2a721d9aecffdfdddc5c83e27f7592594f71932481",
			},
		},
		{
			name: "Number",
			conf: `
i: '=40+2'
`,
			expected: params{
				I: float64(42),
			},
		},
		{
			name: "Struct",
			conf: `
i:
  item1: ="123"
  item2:
  - ="item21"
  - item22
  item3: =789
l1:
  - l11
  - ="l12"
`,
			expected: params{
				I: map[string]interface{}{
					"item1": "123",
					"item2": toInterfaceList([]string{"item21", "item22"}),
					"item3": float64(789),
				},
				L1: []string{"l11", "l12"},
			},
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			conf := getConf(t, c.conf)
			context := ctx.New()
			context.In.Req.Method = "POST"
			context.V = ctx.Variables{
				"variable1": "value1",
				"variable2": []string{"value21", "value22"},
				"variable3": map[string]string{
					"f1": "value31",
					"f2": "value32",
					"f3": "value33",
				},
				"variable4": `{"name": "value4"}`,

				"fuzzy": []string{"cartwheel", "foobar", "wheel", "baz"},
			}
			p := params{}
			if !GetPredicateParams(context, &conf, &p) {
				t.Errorf("Cannot convert params %v", conf)
			} else if diff := deep.Equal(p, c.expected); diff != nil {
				t.Error(diff)
			}
		})

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
func getConf(t *testing.T, source string) Params {
	c := Params{Name: "test"}
	require.Nil(t, yaml.Unmarshal([]byte(source), &(c.Conf)), "YAML parsing error: %v")
	return c
}

func toInterfaceList(l []string) []interface{} {
	r := make([]interface{}, len(l))
	for i, v := range l {
		r[i] = v
	}
	return r
}
