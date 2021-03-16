// ctx defines the context available for predicates, that is, the data
// about the incoming request or the result of the previous evaluated
// predicates
package ctx

import (
	"net/http"
	"net/url"
)

// Main entry point to the context
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

// Info about incoming URL
type Url struct {
	Params url.Values //map[string]string
}

// Info about incoming request
type Request struct {
	// Incoming HTTP request
	Req *http.Request

	// Imcoming URL info
	Url *Url

	// Mime type of body
	Mime string

	// Content of body
	Body string
}

// Results resgistered by a predicate
type Registered map[string]map[string]interface{}

// Variables set by the 'set' option
type Variables map[string]interface{}

// Resulsts of a predicate
type Results map[string]interface{}

// Default predicates values, set by 'default' predicate
type Default map[string]map[string]interface{}

// Convert context to generic type for Gval evaluation
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
