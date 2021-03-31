// Copyright 2021 Jérôme Sautret. All rights reserved.  Use of this
// source code is governed by an Apache License 2.0 that can be found
// in the LICENSE file.

// Package ctx defines the context available for predicates, that is, the data
// about the incoming request or the result of the previous evaluated
// predicates
package ctx

import (
	"net/http"
	"net/url"
)

// Ctx is the main entry point to the context
type Ctx struct {
	// Incoming request
	In *http.Request

	// Default predicates values, set by 'default' predicate
	Default Default

	// Variables set by the 'variable' option
	V Variables

	// Results registered by previous predicates, using the
	// 'register' option
	R Registered

	// Value of last evaluated predicate
	Result bool
}

// New returns a empty context
func New() *Ctx {
	return &Ctx{
		In:      &http.Request{},
		Default: Default{},
		R:       Registered{},
		V:       Variables{},
	}
}

// URL contains info about incoming URL
type URL struct {
	Params url.Values // map[string]string
}

// Registered stores results resgistered by a predicate
type Registered map[string]Result

// Result is the type data returned by predicates
type Result map[string]interface{}

// Variables set by the 'variable' option
type Variables map[string]interface{}

// Default stores predicates values, set by 'default' predicate
type Default map[string]DefaultParams

// DefaultParams stores the default parameters for a predicate type
type DefaultParams map[string]interface{}
