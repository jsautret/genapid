// Functions to convert data

// Contains some code with following license:
// The MIT License (MIT)
//
// Copyright (c) 2014 Heye Vöcking
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package conf

import (
	"context"
	"encoding/json"
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

// Traverse data structure & evaluate Gval expression in strings
func convert(obj interface{}, c *ctx.Ctx) interface{} {
	// Wrap the original in a reflect.Value
	original := reflect.ValueOf(obj)

	copy := reflect.New(original.Type()).Elem()
	convertRecursive(copy, original, c)

	// Remove the reflection wrapper
	return copy.Interface()
}

func convertRecursive(copy, original reflect.Value, c *ctx.Ctx) {
	switch original.Kind() {
	// The first cases handle nested structures and translate them recursively

	// If it is a pointer we need to unwrap and call once again
	case reflect.Ptr:
		// To get the actual value of the original we have to call Elem()
		// At the same time this unwraps the pointer so we don't end up in
		// an infinite recursion
		originalValue := original.Elem()
		// Check if the pointer is nil
		if !originalValue.IsValid() {
			return
		}
		// Allocate a new object and set the pointer to it
		copy.Set(reflect.New(originalValue.Type()))
		// Unwrap the newly created pointer
		convertRecursive(copy.Elem(), originalValue, c)

	// If it is an interface (which is very similar to a pointer), do basically the
	// same as for the pointer. Though a pointer is not the same as an interface so
	// note that we have to call Elem() after creating a new object because otherwise
	// we would end up with an actual pointer
	case reflect.Interface:
		if !original.IsZero() {
			// Get rid of the wrapping interface
			originalValue := original.Elem()
			// Create a new object. Now new gives us a pointer, but we want the value it
			// points to, so we have to call Elem() to unwrap it
			copyValue := reflect.New(originalValue.Type()).Elem()

			if originalValue.Kind() == reflect.String {
				newValue := convertElem(originalValue, c)
				p := reflect.New(original.Type())
				p.Elem().Set(original)

				copy.Set(newValue)
			} else {
				convertRecursive(copyValue, originalValue, c)
				copy.Set(copyValue)
			}
		}

	// If it is a struct we translate each field
	case reflect.Struct:
		for i := 0; i < original.NumField(); i++ {
			if original.Field(i).CanSet() {
				convertRecursive(copy.Field(i), original.Field(i), c)
			}
		}

	// If it is a slice we create a new slice and translate each element
	case reflect.Slice:
		copy.Set(reflect.MakeSlice(original.Type(), original.Len(), original.Cap()))
		for i := 0; i < original.Len(); i++ {
			convertRecursive(copy.Index(i), original.Index(i), c)
		}

	// If it is a map we create a new map and translate each value
	case reflect.Map:
		copy.Set(reflect.MakeMap(original.Type()))
		for _, key := range original.MapKeys() {
			originalValue := original.MapIndex(key)
			// New gives us a pointer, but again we want the value
			copyValue := reflect.New(originalValue.Type()).Elem()
			convertRecursive(copyValue, originalValue, c)
			copy.SetMapIndex(key, copyValue)
		}

	// Otherwise we cannot traverse anywhere so this finishes the the recursion

	// If it is a string convert it
	case reflect.String:
		r := convertElem(original, c)
		if s, ok := r.Interface().(string); ok {
			copy.SetString(s)
		} else {
			copy.Set(original)
		}

	// And everything else will simply be taken from the original
	default:
		copy.Set(original)
	}

}

// Convert Elem if it's a string, remplacing Gval expression by its value
func convertElem(v reflect.Value, c *ctx.Ctx) reflect.Value {
	//fmt.Printf("gval1 %v\n", v)
	if v.Kind() == reflect.String {
		//fmt.Printf("gval2 %v\n", v)
		r, err := evaluateGval(v.String(), c)
		if err != nil {
			log.Warn().Err(err).Msg("Cannot evaluate Gval expression")
			return v
		}
		//fmt.Printf("gval3 %v\n", r)
		//fmt.Printf("gval4 %v\n", r.Kind())
		return reflect.ValueOf(r)
	}
	return v
}

// Evaluate string if it's a Gval expression and return its value
func evaluateGval(s string, c *ctx.Ctx) (interface{}, error) {
	if s != "" && s[0] == '=' {
		return gval.Evaluate(s[1:], c,
			jsonpath.Language(), jsonpathFunction(),
			pipeOperator(), fuzzyFunction())
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
