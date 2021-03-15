package conf

import (
	"net/http"
	"os"
	"testing"

	"github.com/go-test/deep"
	"github.com/jsautret/go-api-broker/ctx"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"
)

var logLevel = zerolog.FatalLevel

//var logLevel = zerolog.TraceLevel

type params struct {
	S1, S2 string
	I      interface{}
	L1     []string
	L2     []interface{}
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
s1: '=Req.Method'
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
			name: "Fuzzy",
			conf: `
s1: '=fuzzy("whl", V.fuzzy)'
`,
			expected: params{
				S1: "wheel",
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
				I: Params{
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
			context := ctx.Ctx{
				Req: &http.Request{Method: "POST"},
				R:   make(ctx.Registered),
			}
			context.V = ctx.Variables{
				"variable1": "value1",
				"variable2": []string{"value21", "value22"},
				"variable3": map[string]string{
					"f1": "value31",
					"f2": "value32",
					"f3": "value33",
				},
				"fuzzy": []string{"cartwheel", "foobar", "wheel", "baz"},
			}
			p := params{}
			if !GetParams(&context, conf, &p) {
				t.Errorf("Cannot convert params %v", conf)
			} else {
				//fmt.Printf("I %v\n", reflect.ValueOf(p.I).Kind())
				//item1 := p2.(*params).I.(Params)["item1"]
				//t.Logf("item1 %v (%v)", item1, reflect.TypeOf(item1))
				//item3 := p2.(*params).I.(Params)["item3"]
				//t.Logf("item3 %v (%v)", item3, reflect.TypeOf(item3))
				if diff := deep.Equal(p, c.expected); diff != nil {
					t.Error(diff)
				}
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
	c := Params{}
	if err := yaml.Unmarshal([]byte(source), &c); err != nil {
		t.Fatalf("YAML parsing error: %v", err)
	}
	return c
}

func getConfB(source string) Params {
	c := Params{}
	yaml.Unmarshal([]byte(source), &c)
	return c
}

func toInterfaceList(l []string) []interface{} {
	r := make([]interface{}, len(l))
	for i, v := range l {
		r[i] = v
	}
	return r
}
