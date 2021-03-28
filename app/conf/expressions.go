// Copyright 2021 Jérôme Sautret. All rights reserved.  Use of this
// source code is governed by an Apache License 2.0 that can be found
// in the LICENSE file.

package conf

import (
	"context"
	"crypto/hmac"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"reflect"
	"sort"
	"strings"

	"github.com/PaesslerAG/gval"
	"github.com/PaesslerAG/jsonpath"
	"github.com/jsautret/genapid/ctx"
	fuzzysearch "github.com/lithammer/fuzzysearch/fuzzy"
	"github.com/rs/zerolog/log"
)

// Evaluate string if it's a Gval expression and return its value
func evaluateGval(s string, c *ctx.Ctx) (interface{}, error) {
	if s != "" && s[0] == '=' {
		return gval.Evaluate(s[1:], c,
			jsonpath.Language(), jsonpathFunction(),
			pipeOperator(), fuzzyFunction(), formatFunction(),
			lenFunction(), upperFunction(), hmacSha256Function(),
			hmacSha1Function())
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
		if len(matches) > 0 {
			sort.Sort(matches)
			return matches[0].Target, nil
		}
		return "", nil
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

// Add a len(list) function to Gval
func lenFunction() gval.Language {
	return gval.Function("len", func(arguments ...interface{}) (interface{}, error) {
		if len(arguments) != 1 {
			return nil, errors.New("len() expects exactly one argument")
		}
		l, ok := arguments[0].([]interface{})
		if !ok {
			return nil, errors.New("len() expects list as argument")
		}
		return len(l), nil
	})
}

// Add a upper(string, parameters...) function to Gval
func upperFunction() gval.Language {
	return gval.Function("upper", func(arguments ...interface{}) (interface{}, error) {
		if len(arguments) != 1 {
			return nil, errors.New("upper() expects exactly one argument")
		}
		s, ok := arguments[0].(string)
		if !ok {
			return nil, errors.New("upper() expects string as argument")
		}
		return strings.ToUpper(s), nil
	})
}

// Add a hmacSha256(string, parameters...) function to Gval
func hmacSha256Function() gval.Language {
	return gval.Function("hmacSha256", func(arguments ...interface{}) (interface{}, error) {
		if len(arguments) != 2 {
			return nil, errors.New("hmacSha256() expects exactly two arguments")
		}
		k, ok := arguments[0].(string)
		if !ok {
			return nil, errors.New("hmacSha256() expects string as arguments")
		}
		d, ok := arguments[1].(string)
		if !ok {
			return nil, errors.New("hmacSha256() expects string as arguments")
		}

		h := hmac.New(sha256.New, []byte(k))
		if _, err := h.Write([]byte(d)); err != nil {
			return nil, errors.New("hmacSha256() cannot write hash")
		}
		return hex.EncodeToString(h.Sum(nil)), nil
	})
}

// Add a hmacSha1(string, parameters...) function to Gval
func hmacSha1Function() gval.Language {
	return gval.Function("hmacSha1", func(arguments ...interface{}) (interface{}, error) {
		if len(arguments) != 2 {
			return nil, errors.New("hmacSha1() expects exactly two arguments")
		}
		k, ok := arguments[0].(string)
		if !ok {
			return nil, errors.New("hmacSha1() expects string as arguments")
		}
		d, ok := arguments[1].(string)
		if !ok {
			return nil, errors.New("hmacSha1() expects string as arguments")
		}

		h := hmac.New(sha1.New, []byte(k))
		if _, err := h.Write([]byte(d)); err != nil {
			return nil, errors.New("hmacSha1() cannot write hash")
		}
		return hex.EncodeToString(h.Sum(nil)), nil
	})
}
