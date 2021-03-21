package logpredicate

import (
	"fmt"

	"github.com/jsautret/go-api-broker/ctx"
	"github.com/jsautret/go-api-broker/genapid"
	"github.com/rs/zerolog"
)

// Name of the predicate
var Name = "log"

// Predicate is the conf.Plugin interface that describes the predicate
type Predicate struct {
	name   string
	params struct { // Params accepted by the predicate
		Msg interface{} `validate:"required"`
	}
}

// Call evaluate a predicate
func (predicate *Predicate) Call(log zerolog.Logger) bool {
	log.Info().Str("log", fmt.Sprintf("%v", predicate.params.Msg)).Msg("")

	return true
}

// Generic interface //

// Result returns data set by the predicate
func (predicate *Predicate) Result() ctx.Result {
	// no data is set by log
	return ctx.Result{}
}

// Name returns the name of the predicate
func (predicate *Predicate) Name() string {
	return predicate.name
}

// Params returns a reference to the params struct of the predicate
func (predicate *Predicate) Params() interface{} {
	return &predicate.params
}

// New returns a new Predicate
func New() genapid.Predicate {
	return &Predicate{
		name: Name,
	}
}
