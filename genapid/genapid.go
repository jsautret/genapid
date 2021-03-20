package genapid

//go:generate mockery --disable-version-string --log-level error --name Predicate

import (
	"errors"

	"github.com/jsautret/go-api-broker/ctx"
	"github.com/jsautret/go-api-broker/internal/conf"
	"github.com/rs/zerolog"
)

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
	if !conf.GetPredicateParams(c, cfg, p.Params()) {
		log.Error().Err(errors.New("Invalid params")).Msg("")
		return false
	}
	return true
}
