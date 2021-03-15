package pipe

import (
	"net/http"

	"github.com/jsautret/go-api-broker/ctx"
	"github.com/jsautret/go-api-broker/internal/conf"
	"github.com/jsautret/go-api-broker/internal/predicate"
	"github.com/rs/zerolog/log"
)

var defaultPipeName = "noname"

func Process(p conf.Pipe, r *http.Request) bool {
	name := p.Name
	if name == "" {
		name = defaultPipeName
	}
	log := log.With().Str("pipe", name).Logger()

	log.Debug().Msgf("Processing pipe '%v'", name)
	url := &ctx.Url{
		Params: r.URL.Query(),
	}
	ctx := &ctx.Ctx{
		Req: r,
		Url: url,
		R:   make(ctx.Registered),
		V:   make(ctx.Variables),
	}

	var result bool
	for j := 0; j < len(p.Pipe); j++ {
		result = predicate.Process(p.Pipe[j], ctx, r)
		if !result {
			break
		}

	}
	log.Debug().Bool("value", result).Msg("End pipe")
	return result
}
