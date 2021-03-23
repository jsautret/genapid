package genapid

//go:generate mockery --disable-version-string --log-level error --name Predicate

import (
	"context"
	"errors"
	"reflect"

	"github.com/go-playground/mold/v4"
	"github.com/go-playground/mold/v4/modifiers"
	"github.com/go-playground/validator/v10"
	"github.com/jsautret/go-api-broker/ctx"
	"github.com/jsautret/go-api-broker/internal/conf"
	"github.com/jsautret/go-api-broker/internal/fileutils"
	"github.com/rs/zerolog"
)

var (
	// use a single instance of Validate, it caches struct info
	validate = validator.New()
	modify   = modifiers.New()
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
	params := p.Params()
	if !conf.GetPredicateParams(c, cfg, params) {
		log.Error().Err(errors.New("Invalid params")).Msg("")
		return false
	}

	if t := reflect.ValueOf(params); t.Kind() == reflect.Ptr &&
		t.Elem().Kind() == reflect.Struct {
		if err := modify.Struct(context.Background(), params); err != nil {
			log.Error().Err(err).Msg("")
			return false
		}

		if err := validate.Struct(params); err != nil {
			log.Error().Err(err).Msg("")
			return false
		}
	}
	return true
}

func init() {
	modify.Register("path", pathModifier)
}

// Expands ~ like the shell would do
func pathModifier(ctx context.Context, fl mold.FieldLevel) error {
	s, ok := fl.Field().Interface().(string)
	if !ok {
		return nil
	}
	fl.Field().SetString(fileutils.Path(s))
	return nil
}
