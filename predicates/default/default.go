package defaultpredicate

import (
	"github.com/jsautret/go-api-broker/ctx"
	"github.com/jsautret/go-api-broker/internal/conf"
)

// Name returns the name the predicate
func (Predicate) Name() string {
	return "default"
}

// Get returns the plugin for the default predicate
func Get() Predicate {
	return Predicate{}
}

// Predicate implements the conf.Plugin interface
type Predicate struct{}

// Call evaluate the predicate
func (Predicate) Call(ctx *ctx.Ctx, config conf.Params) bool {
	conf.AddDefault(ctx, config)

	return true
}
