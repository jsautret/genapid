package chromecastpredicate

import (
	"io/ioutil"
	"os"

	"github.com/jsautret/go-api-broker/ctx"
	"github.com/jsautret/go-api-broker/genapid"
	"github.com/jsautret/go-api-broker/internal/fileutils"
	"github.com/rs/zerolog"
	"github.com/vishen/go-chromecast/tts"
)

// Name of the predicate
var Name = "chromecast"

// Predicate is the conf.Plugin interface that describes the predicate
type Predicate struct {
	name   string
	params struct { // Params accepted by the predicate
		GoogleServiceAccount string  `validate:"required" mapstructure:"google_service_account"`
		LanguageCode         string  `validate:"required" mapstructure:"language_code"`
		VoiceName            string  `validate:"required" mapstructure:"voice_name"`
		Addr                 string  `validate:"required,ip"`
		Port                 int     `validate:"required" mod:"default=8009"`
		TTS                  string  `validate:"required"`
		SpeakingRate         float32 `validate:"required" mod:"default=1.0" mapstructure:"speaking_rate"`
		Pitch                float32 `validate:"required" mod:"default=1.0"`
	}
}

// Call evaluates the predicate
func (predicate *Predicate) Call(log zerolog.Logger) bool {
	p := predicate.params
	log.Debug().Str("tts", p.TTS).Msg("")

	b, err := ioutil.ReadFile(fileutils.Path(p.GoogleServiceAccount))
	if err != nil {
		log.Error().Err(err).Msg("Unable to open google service account file")
		return false
	}
	app, err := castApplication(p.Addr, p.Port)
	if err != nil {
		log.Error().Err(err).Msg("unable to get cast application")
		return false
	}

	data, err := tts.Create(p.TTS, b, p.LanguageCode, p.VoiceName, p.SpeakingRate, p.Pitch)
	if err != nil {
		log.Error().Err(err).Msg("")
		return false
	}
	f, err := ioutil.TempFile("", "go-chromecast-tts")
	if err != nil {
		log.Error().Err(err).Msg("Unable to create temp file")
		return false
	}

	if _, err := f.Write(data); err != nil {
		log.Error().Err(err).Msg("Unable to write to temp file")
		return false
	}

	if err := app.Load(f.Name(), "audio/mp3", false, false, false); err != nil {
		log.Error().Err(err).Msg("unable to load media to device")
		return false
	}

	if err := os.Remove(f.Name()); err != nil {
		log.Error().Err(err).Msg("unable clean up temp file")
		// we don't fail the predicate for that
	}

	return true
}

// Generic interface //

// Result returns data set by the predicate
func (predicate *Predicate) Result() ctx.Result {
	return ctx.Result{}
}

// Name returns the name of the predicate
func (predicate *Predicate) Name() string {
	return predicate.name
}

// Params returns a reference to the params struct of the predicate
func (predicate *Predicate) Params() interface{} {
	return &predicate.params
}

// New returns a new Predicate
func New() genapid.Predicate {
	return &Predicate{
		name: Name,
	}
}
