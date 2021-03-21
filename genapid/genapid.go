package genapid

//go:generate mockery --disable-version-string --log-level error --name Predicate

import (
	"errors"

	"github.com/go-playground/validator/v10"
	"github.com/jsautret/go-api-broker/ctx"
	"github.com/jsautret/go-api-broker/internal/conf"
	"github.com/rs/zerolog"
)

// use a single instance of Validate, it caches struct info
var validate *validator.Validate

// Predicate is the interface of a predicate plugin
type Predicate interface {
	Name() string
	Call(zerolog.Logger) bool
	Result() ctx.Result
	Params() interface{}
}

// InitPredicate sets the parameters from the conf
func InitPredicate(log zerolog.Logger, c *ctx.Ctx,
	p Predicate, cfg *conf.Params) bool {
	params := p.Params()
	if !conf.GetPredicateParams(c, cfg, params) {
		log.Error().Err(errors.New("Invalid params")).Msg("")
		return false
	}
	if err := validate.Struct(params); err != nil {
		if _, ok := err.(*validator.InvalidValidationError); ok {
			// We are probably in a test simulating a param struct
			return true
		}
		log.Error().Err(err).Msg("")
		return false
	}
	return true
}

func init() {
	validate = validator.New()
}
