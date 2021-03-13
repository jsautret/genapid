package main

import (
	"net/http"
	"os"

	"github.com/jsautret/go-api-broker/internal/conf"
	"github.com/jsautret/go-api-broker/internal/pipe"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var (
	config conf.Root
)

func quit(w http.ResponseWriter, r *http.Request) {
	log.Info().Str("app", "stopped").Msg("Application stopped")
	os.Exit(0)
}

func handler(w http.ResponseWriter, r *http.Request) {
	log.Debug().Str("http", "start").Str("path", r.URL.Path).
		Msg("Processing HTTP request")
	var res bool
	for i := 0; i < len(config); i++ {
		res = pipe.Process(config[i], r)
	}
	log.Debug().Str("http", "end").Str("path", r.URL.Path).
		Msg("HTTP request processed")
	if !res {
		w.WriteHeader(http.StatusNotFound)
	}
}

func main() {
	config = conf.Read("route.yml")
	http.HandleFunc("/quit", quit)
	http.HandleFunc("/", handler)

	log.Info().Str("app", "started").Int("port", 9191).
		Msgf("Application started and listening to :%v", 9191)

	log.Fatal().Err(http.ListenAndServe(":9191", nil)).Msg("")
}

func init() {
	// UNIX Time is faster and smaller than most timestamps
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	//log.Logger = zerolog.New(os.Stdout).With().Timestamp().Logger()
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
}
