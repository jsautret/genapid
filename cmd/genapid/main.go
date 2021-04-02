// Copyright 2021 Jérôme Sautret. All rights reserved.  Use of this
// source code is governed by an Apache License 2.0 that can be found
// in the LICENSE file.

package main

import (
	"flag"
	"net/http"
	"os"
	"strconv"

	"github.com/jsautret/genapid/app/conf"
	"github.com/jsautret/genapid/app/plugins"
	"github.com/jsautret/genapid/ctx"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var (
	// Global conf
	config conf.Root
	// main context
	staticCtx *ctx.Ctx
)

// Command line flags variables
var (
	configFileName string
	SLogLevel      string
	port           int
)

// Command line flags definitions
func init() {
	flag.StringVar(&configFileName, "config", "api.yml", "Config file")
	flag.StringVar(&SLogLevel, "loglevel", "info", "Log level")
	flag.IntVar(&port, "port", 9110, "Listening port")
}

// Main handler for incoming requests
func handler(w http.ResponseWriter, r *http.Request) {
	process(w, r, staticCtx)
}

func main() {
	flag.Parse()

	// UNIX Time is faster and smaller than most timestamps
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	if fileInfo, _ := os.Stdout.Stat(); (fileInfo.Mode() & os.ModeCharDevice) != 0 {
		// sdtout is console
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	}
	if logLevel, err := zerolog.ParseLevel(SLogLevel); err != nil {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
		log.Warn().Err(err).Msg("Forcing info log level")
	} else {
		zerolog.SetGlobalLevel(logLevel)
		if logLevel == zerolog.TraceLevel {
			log.Logger = log.With().Caller().Timestamp().Logger()
		}
		log.Info().Str("loglevel", logLevel.String()).Msg("Setting loglevel")
	}

	config = conf.ReadConfFile(configFileName)
	staticCtx = ctx.New()
	processInit(&config, staticCtx)

	for k := range plugins.List() {
		log.Info().Str("plugin", k).Msg("Plugin enabled")
	}

	server := http.NewServeMux()
	server.HandleFunc("/", handler)
	log.Info().Str("app", "started").Int("port", port).
		Msgf("Application started and listening to :%v", port)

	log.Fatal().Err(http.ListenAndServe(":"+strconv.Itoa(port), server)).Msg("")
}
