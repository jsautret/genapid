// Copyright 2021 Jérôme Sautret. All rights reserved.  Use of this
// source code is governed by an Apache License 2.0 that can be found
// in the LICENSE file.

package main

import (
	"net/http"

	"github.com/jsautret/genapid/app/predicate"
	"github.com/jsautret/genapid/ctx"
	"github.com/rs/zerolog/log"
)

// Process incoming request
func process(w http.ResponseWriter, r *http.Request) bool {
	log.Debug().Str("http", "start").Str("path", r.URL.Path).
		Msg("Processing HTTP request")

	log.Trace().Interface("headers", r.Header).Msg("")

	// Create & init context structures
	url := &ctx.URL{
		Params: r.URL.Query(),
	}
	In := ctx.Request{
		Req: r,
		URL: url,
	}
	c := ctx.New()
	c.In = In

	// Process each pipe
	var res bool
	for i := 0; i < len(config); i++ {
		pc := config[i]
		res = predicate.ProcessPipe(&pc, c)
	}
	log.Debug().Str("http", "end").Str("path", r.URL.Path).
		Msg("HTTP request processed")
	// return result of last predicate in pipe
	// used for tests
	return res
}
