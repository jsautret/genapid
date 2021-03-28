// Copyright 2021 Jérôme Sautret. All rights reserved.  Use of this
// source code is governed by an Apache License 2.0 that can be found
// in the LICENSE file.

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
	"reflect"

	"github.com/jsautret/genapid/ctx"
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
	if v.Kind() == reflect.String {
		r, err := evaluateGval(v.String(), c)
		if err != nil {
			log.Warn().Err(err).Msg("Cannot evaluate Gval expression")
			return v
		}
		return reflect.ValueOf(r)
	}
	return v
}
