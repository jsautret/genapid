package defaultPredicate

import (
	"github.com/jsautret/go-api-broker/ctx"
	"github.com/jsautret/go-api-broker/internal/conf"
)

func Call(ctx *ctx.Ctx, config conf.Params) bool {
	conf.AddDefault(ctx, config)

	return true
}
