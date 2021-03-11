package context

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
	R map[string]map[string]interface{}

	// Return
	Results map[string]interface{}
}

type Url struct {
	Params url.Values //map[string]string
}
