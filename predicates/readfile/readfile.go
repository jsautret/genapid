package readfilepredicate

import (
	"encoding/json"
	"io/ioutil"

	"github.com/jsautret/go-api-broker/ctx"
	"github.com/jsautret/go-api-broker/genapid"
	"github.com/jsautret/go-api-broker/internal/fileutils"
	"github.com/rs/zerolog"
	"gopkg.in/yaml.v3"
)

// Name of the predicate
var Name = "readfile"

// Predicate is the conf.Plugin interface that describes the predicate
type Predicate struct {
	name   string
	params struct { // Params accepted by the predicate
		JSON string `validate:"required_without=YAML,excluded_with=YAML"`
		YAML string `validate:"required_without=JSON,excluded_with=JSON"`
	}
	results ctx.Result // content of file
}

// Call evaluates the predicate
func (predicate *Predicate) Call(log zerolog.Logger) bool {
	p := predicate.params
	if p.YAML != "" {
		log.Debug().Str("yaml", p.YAML).Msg("Reading file")
		y, err := ioutil.ReadFile(fileutils.Path(p.YAML))
		if err != nil {
			log.Error().Err(err).Msg("")
			return false
		}
		var result interface{}
		if err := yaml.Unmarshal(y, &result); err != nil {
			log.Error().Err(err).Msg("Invalid YAML")
			return false
		}
		predicate.results = ctx.Result{"content": result}
		return true
	}
	if p.JSON != "" {
		log.Debug().Str("json", p.JSON).Msg("Reading file")
		j, err := ioutil.ReadFile(fileutils.Path(p.JSON))
		if err != nil {
			log.Error().Err(err).Msg("")
			return false
		}
		var result interface{}
		if err := json.Unmarshal(j, &result); err != nil {
			log.Error().Err(err).Msg("Invalid JSON")
			return false
		}
		predicate.results = ctx.Result{"content": result}
		return true
	}
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
