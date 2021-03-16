package main

import (
	"io/ioutil"
	"mime"
	"net/http"

	"github.com/jsautret/go-api-broker/ctx"
	"github.com/jsautret/go-api-broker/internal/predicate"
	"github.com/rs/zerolog/log"
)

// Process incoming request
func process(w http.ResponseWriter, r *http.Request) bool {
	log.Debug().Str("http", "start").Str("path", r.URL.Path).
		Msg("Processing HTTP request")

	var err error

	// Create & init context structures
	url := &ctx.Url{
		Params: r.URL.Query(),
	}
	contentType := r.Header.Get("Content-type")
	m := ""
	m, _, err = mime.ParseMediaType(contentType)
	if err != nil {
		log.Debug().Err(err).Msg("Ignoring Content-type")
	} else {
		log.Debug().Str("mime", m).Msg("Found Content-type")
	}
	var body []byte
	if body, err = ioutil.ReadAll(r.Body); err != nil {
		log.Debug().Err(err).Msg("Error reading body")
	}
	In := ctx.Request{
		Req:  r,
		Url:  url,
		Mime: m,
		Body: string(body),
	}
	c := ctx.Ctx{
		In:      In,
		R:       make(ctx.Registered),
		V:       make(ctx.Variables),
		Default: make(ctx.Default),
	}

	// Process each pipe
	var res bool
	for i := 0; i < len(config); i++ {
		res = predicate.ProcessPipe(config[i], &c)
	}
	log.Debug().Str("http", "end").Str("path", r.URL.Path).
		Msg("HTTP request processed")
	// return result of last predicate in pipe
	// used for tests
	return res
}
