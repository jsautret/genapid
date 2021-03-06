// Copyright 2021 Jérôme Sautret. All rights reserved.  Use of this
// source code is governed by an Apache License 2.0 that can be found
// in the LICENSE file.

package predicate

import (
	"github.com/jsautret/genapid/app/conf"
	"github.com/jsautret/genapid/ctx"
	"github.com/rs/zerolog"
)

// ProcessPipe evaluate a pipe of predicates
func ProcessPipe(log zerolog.Logger, p *conf.Pipe, c *ctx.Ctx) bool {
	name := p.Name
	log = log.With().Str("pipe", name).Logger()

	log.Debug().Msgf("Processing pipe '%v'", name)

	// save defaults
	d := copyDefault(c.Default)
	var result bool
	for j := 0; j < len(p.Pipe); j++ {
		result = Process(log, &p.Pipe[j], c)
		if !result {
			break
		}

	}
	// restore defaults
	c.Default = d
	log.Debug().Bool("value", result).Msg("End pipe")
	// return result of last predicate
	// used for tests
	return result
}

func copyDefault(d ctx.Default) ctx.Default {
	n1 := ctx.Default{}
	for k1, v1 := range d {
		n2 := ctx.DefaultParams{}
		for k2, v2 := range v1 {
			n2[k2] = v2
		}
		n1[k1] = n2
	}

	return n1
}
