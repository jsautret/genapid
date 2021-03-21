package conf

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"reflect"
	"sort"
	"strings"

	"github.com/PaesslerAG/gval"
	"github.com/PaesslerAG/jsonpath"
	"github.com/jsautret/go-api-broker/ctx"
	fuzzysearch "github.com/lithammer/fuzzysearch/fuzzy"
	"github.com/rs/zerolog/log"
)

// Evaluate string if it's a Gval expression and return its value
func evaluateGval(s string, c *ctx.Ctx) (interface{}, error) {
	if s != "" && s[0] == '=' {
		return gval.Evaluate(s[1:], c,
			jsonpath.Language(), jsonpathFunction(),
			pipeOperator(), fuzzyFunction(), formatFunction())
	}
	return s, nil
}

// Add a pipe operator to Gval expressions to pass a new Gval context to
// another Gval expression
func pipeOperator() gval.Language {
	return gval.PostfixOperator("|",
		func(c context.Context, p *gval.Parser,
			pre gval.Evaluable) (gval.Evaluable, error) {
			post, err := p.ParseExpression(c)
			if err != nil {
				return nil, err
			}
			return func(c context.Context,
				v interface{}) (interface{}, error) {
				v, err := pre(c, v)
				if err != nil {
					return nil, err
				}
				return post(c, v)
			}, nil
		})
}

// Add a jsonpath(path, json) function to Gval that returns the
// evaluated json path on the json data. Data will be converted to
// json if possible.
func jsonpathFunction() gval.Language {
	return gval.Function("jsonpath", func(arguments ...interface{}) (interface{}, error) {
		if len(arguments) != 2 {
			return nil, fmt.Errorf("jsonpath() expects exactly two arguments")
		}
		path, ok := arguments[0].(string)
		if !ok {
			return nil, fmt.Errorf("jsonpath() expects string as first argument")
		}
		fmt.Println("XXX", arguments[0])
		if r, err := toJSON(arguments[1]); err != nil {
			log.Warn().Err(err).Msg("Cannot convert to json")
		} else {
			return jsonpath.Get(path, r)
		}
		return jsonpath.Get(path, arguments[1])

	})
}

// ToJSON tries to convert data to something that jsonpath can understand
func toJSON(in interface{}) (map[string]interface{}, error) {
	log := log.With().Str("jsonpath", "toJson").Logger()
	log.Trace().Interface("in", in).Msg("")
	switch reflect.TypeOf(in).Kind() {
	case reflect.Map:
		log.Trace().Str("type", "map").Msg("")
		out := make(map[string]interface{})
		iter := reflect.ValueOf(in).MapRange()
		for iter.Next() {
			k := iter.Key()
			v := iter.Value()
			out[k.String()] = v.Interface()
		}
		return out, nil
	case reflect.Struct:
		log.Trace().Str("type", "struct").Msg("")
		if r, ok := in.(io.Reader); ok {
			var v interface{}

			err := json.NewDecoder(r).Decode(&v)
			if err != nil {
				return nil, err
			}
			return toJSON(v)
		}
	case reflect.String:
		log.Trace().Str("type", "string").Msg("")
		var v interface{}
		s, _ := in.(string)
		err := json.NewDecoder(strings.NewReader(s)).Decode(&v)
		if err != nil {
			return nil, err
		}
		log.Trace().Interface("in", in).Msg("")
		return toJSON(v)
	}

	err := fmt.Errorf("cannot convert to json: %v", in)
	return nil, err
}

// Add a fuzzy(string, stringList) function to Gval that returns the
// best fuzzy match from the list
func fuzzyFunction() gval.Language {
	return gval.Function("fuzzy", func(arguments ...interface{}) (interface{}, error) {
		if len(arguments) != 2 {
			return nil,
				fmt.Errorf("fuzzy() expects exactly two arguments")
		}
		source, ok := arguments[0].(string)
		if !ok {
			return nil,
				fmt.Errorf("fuzzy() expects string as first argument")
		}
		targets, err := toListStrings(arguments[1])
		if err != nil {
			return nil,
				fmt.Errorf("fuzzy() expects []string as second argument, %v", err)
		}
		matches := fuzzysearch.RankFindNormalizedFold(
			source, targets)
		sort.Sort(matches)
		return matches[0].Target, nil
	})
}

// Tries to convert data to a list of strings
func toListStrings(arg interface{}) ([]string, error) {
	if r, ok := arg.([]string); ok {
		return r, nil
	}
	if reflect.TypeOf(arg).Kind() == reflect.Slice {
		l := arg.([]interface{})
		ls := make([]string, len(l))
		for i := 0; i < len(l); i++ {
			e, ok := l[i].(string)
			if !ok {
				err := fmt.Errorf("element is not string: %v", l[i])
				return nil, err
			}
			ls[i] = e
		}
		return ls, nil
	}
	err := fmt.Errorf("not a list: %v", arg)
	return nil, err
}

// Add a format(string, parameters...) function to Gval
func formatFunction() gval.Language {
	return gval.Function("format", func(arguments ...interface{}) (interface{}, error) {
		if len(arguments) == 0 {
			return nil, errors.New("format() expects at least one argument")
		}
		s, ok := arguments[0].(string)
		if !ok {
			return nil, errors.New("format() expects string as first argument")
		}
		return fmt.Sprintf(s, arguments[1:]...), nil
	})
}
