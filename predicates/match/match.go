package matchpredicate

import (
	"errors"
	"regexp"

	"github.com/jsautret/go-api-broker/ctx"
	"github.com/jsautret/go-api-broker/genapid"
	"github.com/rs/zerolog"
)

// Name of the predicate
var Name = "match"

// Predicate is the conf.Plugin interface that describes the predicate
type Predicate struct {
	name   string
	params struct { // Params accepted by the predicate
		String string `validate:"required"`
		Fixed  string `validate:"required_without=Regexp"`
		Regexp string `validate:"required_without=Fixed"`
	}
	results ctx.Result // result of regexp match
}

// Call evaluate the predicate
func (predicate *Predicate) Call(log zerolog.Logger) bool {
	p := predicate.params
	log.Debug().Str("string", p.String).Msg("")
	if p.Fixed != "" {
		log.Debug().Str("fixed", p.Fixed).Msg("")
		return p.Fixed == p.String
	}
	if p.Regexp != "" {
		log.Debug().Str("regexp", p.Regexp).Msg("")

		r, err := regexp.Compile(p.Regexp)
		if err != nil {
			log.Error().Err(err).Msg("invalid 'regexp'")
			return false
		}
		// get list of matches
		res := r.FindStringSubmatch(p.String)
		if len(res) == 0 {
			return false
		}
		// get named matches
		namedRes := make(map[string]string)
		for i, name := range r.SubexpNames() {
			if i != 0 && name != "" {
				namedRes[name] = res[i]
			}
		}
		log.Debug().Msgf("'regexp' matched %v", res)
		predicate.results = ctx.Result{
			"matches": res,
			"named":   namedRes,
		}
		return true
	}
	// validate should prevent reaching this point
	log.Error().Err(errors.New("Missing one of 'fixed' or 'regexp'")).Msg("")
	return false
}

// Generic interface //

// Result returns data set by the predicate
func (predicate *Predicate) Result() ctx.Result {
	return predicate.results
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
