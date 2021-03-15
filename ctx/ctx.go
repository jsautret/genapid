package ctx

import (
	"net/http"
	"net/url"
)

/* Interface types for predicates */

type Ctx struct {
	// Value of evaluated predicate
	Result bool

	// Incoming HTTP request
	Req *http.Request

	// Imcoming URL info
	Url *Url

	// Registered Contexts
	R Registered

	// Variables
	V Variables

	// Return
	Results Results
}

type Url struct {
	Params url.Values //map[string]string
}
type Registered map[string]map[string]interface{}
type Variables map[string]interface{}
type Results map[string]interface{}

func (c *Ctx) ToInterface() interface{} {
	type ctx struct {
		Req     *http.Request
		Url     *Url
		R       map[string]map[string]interface{}
		V       map[string]interface{}
		Results map[string]interface{}
	}
	return ctx{
		Req:     c.Req,
		Url:     c.Url,
		R:       map[string]map[string]interface{}(c.R),
		V:       map[string]interface{}(c.V),
		Results: map[string]interface{}(c.Results),
	}
}
