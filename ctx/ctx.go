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
	In Request

	// Default predicates values, set by 'default' predicate
	Default Default

	// Variables set by the 'set' option
	V Variables

	// Values set by the latest evaluated predicate
	Results Results

	// Results registered by previous predicates, using the
	// 'register' option
	R Registered

	// Value of last evaluated predicate
	Result bool
}

// URL contains info about incoming URL
type URL struct {
	Params url.Values //map[string]string
}

// Request containas info about incoming request
type Request struct {
	// Incoming HTTP request
	Req *http.Request

	// Imcoming URL info
	URL *URL

	// Mime type of body
	Mime string

	// Content of body
	Body string
}

// Registered stores results resgistered by a predicate
type Registered map[string]map[string]interface{}

// Variables set by the 'set' option
type Variables map[string]interface{}

// Results of a predicate
type Results map[string]interface{}

// Default predicates values, set by 'default' predicate
type Default map[string]map[string]interface{}

// ToInterface converts context to generic type for Gval evaluation
func (c *Ctx) ToInterface() interface{} {
	type ctx struct {
		In      Request
		R       map[string]map[string]interface{}
		V       map[string]interface{}
		Results map[string]interface{}
	}
	return ctx{
		In:      c.In,
		R:       map[string]map[string]interface{}(c.R),
		V:       map[string]interface{}(c.V),
		Results: map[string]interface{}(c.Results),
	}
}
