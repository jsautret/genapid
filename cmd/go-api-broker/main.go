package main

import (
	"flag"
	"net/http"
	"os"

	"github.com/jsautret/go-api-broker/internal/conf"
	"github.com/jsautret/go-api-broker/internal/plugins"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// Global conf
var (
	LogLevel zerolog.Level
	config   conf.Root
)

// Command line flags variables
var (
	configFileName string
	SLogLevel      string
)

// Command line flags definitions
func init() {
	flag.StringVar(&configFileName, "config", "route.yml", "Config file")
	flag.StringVar(&SLogLevel, "loglevel", "info", "Log level")
}

// Main handler for incoming requests
func handler(w http.ResponseWriter, r *http.Request) {
	if !process(w, r) {
		w.WriteHeader(http.StatusNotFound)
	}
}

func quit(w http.ResponseWriter, r *http.Request) {
	log.Info().Str("app", "stopped").Msg("Application stopped")
	os.Exit(0)
}

func main() {
	flag.Parse()

	// UNIX Time is faster and smaller than most timestamps
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
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

	config = conf.Read(configFileName)

	http.HandleFunc("/quit", quit)
	http.HandleFunc("/", handler)

	for k := range plugins.List() {
		log.Info().Str("plugin", k).Msg("Plugin enabled")
	}
	log.Info().Str("app", "started").Int("port", 9191).
		Msgf("Application started and listening to :%v", 9191)

	log.Fatal().Err(http.ListenAndServe(":9191", nil)).Msg("")
}
