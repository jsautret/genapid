package headerpredicate

import (
	"github.com/jsautret/genapid/ctx"
	"github.com/jsautret/genapid/genapid"
	"github.com/rs/zerolog"
)

// Name of the predicate
var Name = "header"

// Predicate is a genapid.Predicate interface that describes the predicate
type Predicate struct {
	name   string
	params struct { // Params accepted by the predicate
		Name  string `validate:"required"`
		Value string
	}
}

// Call evaluates the predicate
func (predicate *Predicate) Call(log zerolog.Logger, c *ctx.Ctx) bool {
	p := predicate.params

	return c.In.Req.Header.Get(p.Name) == p.Value
}

// Generic interface //

// Result returns data set by the predicate
func (predicate *Predicate) Result() ctx.Result {
	return ctx.Result{}
}

// Name returns the name of the predicate
func (predicate *Predicate) Name() string {
	return predicate.name
}

// Params returns a reference to a struct params accepted by the predicate
func (predicate *Predicate) Params() interface{} {
	return &predicate.params
}

// New returns a new Predicate
func New() genapid.Predicate {
	return &Predicate{
		name: Name,
	}
}
