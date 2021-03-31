// Copyright 2021 Jérôme Sautret. All rights reserved.  Use of this
// source code is governed by an Apache License 2.0 that can be found
// in the LICENSE file.

package main

import (
	"errors"

	"github.com/jsautret/genapid/app/conf"
	"github.com/jsautret/genapid/app/predicate"
	"github.com/jsautret/genapid/ctx"
	"github.com/rs/zerolog/log"
)

func processInit(cfg *conf.Root, c *ctx.Ctx) {
	if len(*cfg) == 0 {
		return
	}
	i := (*cfg)[0]
	if i.Init == nil {
		return
	}
	*cfg = (*cfg)[1:] // remove init from the conf
	if i.Pipe != nil {
		log.Error().Err(
			errors.New("'pipe' cannot be used in 'init' section")).Msg("")
		return
	}
	log.Info().Msg("Processing init")
	for j := 0; j < len(i.Init); j++ {
		result := predicate.Process(log.Logger, &i.Init[j], c)
		if !result {
			break
		}
	}
}
