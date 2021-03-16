package defaultpredicate

import (
	"github.com/jsautret/go-api-broker/ctx"
	"github.com/jsautret/go-api-broker/internal/conf"
)

// Call evaluate the predicate
func Call(ctx *ctx.Ctx, config conf.Params) bool {
	conf.AddDefault(ctx, config)

	return true
}
