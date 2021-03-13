package pipe

import (
	"net/http"

	"github.com/jsautret/go-api-broker/context"
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
	url := &context.Url{
		Params: r.URL.Query(),
	}
	ctx := &context.Ctx{
		Req: r,
		Url: url,
		R:   make(context.Registered),
		V:   make(context.Variables),
	}

	var result bool
	for j := 0; j < len(p.Pipe); j++ {
		result = predicate.Process(p.Pipe[j], ctx, r)
		log.Debug().Bool("value", result).Msgf("Predicate is %v", result)
		if !result {
			break
		}

	}
	return result
}
