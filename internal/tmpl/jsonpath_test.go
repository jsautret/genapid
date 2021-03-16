package tmpl

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/go-test/deep"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var logLevel = zerolog.FatalLevel

//var logLevel = zerolog.DebugLevel

var doc = `
{ "store": {
    "book": [ 
      { "category": "reference",
        "author": "Nigel Rees",
        "title": "Sayings of the Century",
        "price": 8.95
      },
      { "category": "fiction",
        "author": "Evelyn Waugh",
        "title": "Sword of Honour",
        "price": 12.99
      },
      { "category": "fiction",
        "author": "Herman Melville",
        "title": "Moby Dick",
        "isbn": "0-553-21311-3",
        "price": 8.99
      },
      { "category": "fiction",
        "author": "J. R. R. Tolkien",
        "title": "The Lord of the Rings",
        "isbn": "0-395-19395-8",
        "price": 22.99
      }
    ],
    "bicycle": {
      "color": "red",
      "price": 19.95
    }
  }
}`

func TestJsonpath(t *testing.T) {
	cases := []struct {
		name     string
		path     string
		expected string
		json     string
	}{
		{
			name:     "Authors",
			expected: `["Nigel Rees", "Evelyn Waugh", "Herman Melville", "J. R. R. Tolkien"]`,
			path:     `$.store.book[*].author`,
			json:     doc,
		},
		{
			name:     "",
			expected: `["Sayings of the Century", "Moby Dick"]`,
			path:     `$..book[?(@.price<10)].title`,
			json:     doc,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			r := jsonpath(c.path, getJSON(t, c.json))
			e := getJSON(t, c.expected)
			if diff := deep.Equal(r, e); diff != nil {
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
func getJSON(t *testing.T, in string) interface{} {
	var res interface{}
	err := json.Unmarshal([]byte(in), &res)
	if err != nil {
		t.Errorf("Cannot parse json")
	}
	return res
}
