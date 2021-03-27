package bodypredicate

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"mime"

	"github.com/jsautret/genapid/ctx"
	"github.com/jsautret/genapid/genapid"
	"github.com/rs/zerolog"
)

// Name of the predicate
var Name = "body"

// Predicate is a genapid.Predicate interface that describes the predicate
type Predicate struct {
	name   string
	params struct { // Params accepted by the predicate
		Type  string `validate:"isdefault|oneof=json string" mod:"lcase"`
		Mime  string
		Limit int64 `mod:"default=1232896"` // 1 Mb
	}
	results ctx.Result // parsed body of incoming request
}

// Call evaluates the predicate
func (predicate *Predicate) Call(log zerolog.Logger, c *ctx.Ctx) bool {
	p := predicate.params

	log.Debug().Str("Type", p.Type).Msg("")
	log.Debug().Str("Mime", p.Mime).Msg("")

	switch c.In.Req.Method {
	case "GET":
		return false
	case "HEAD":
		return false
	case "DELETE":
		return false
	}

	if p.Mime != "" { // checking if content-type match
		contentType := c.In.Req.Header.Get("Content-Type")
		m, _, err := mime.ParseMediaType(contentType)
		if err != nil {
			log.Warn().Err(err).Msg("Cannot parse Content-Type")
			return false
		}
		if m != p.Mime {
			log.Debug().Msgf("Content-type %v doesn't match expected %v", m, p.Mime)
			return false
		}
	}
	if p.Type != "" {
		limitedReader := &io.LimitedReader{R: c.In.Req.Body, N: p.Limit}
		// Use a TeeReade to allow to use body predicate several times
		var b bytes.Buffer
		tee := io.TeeReader(limitedReader, &b)
		switch p.Type {
		case "string":
			body, err := ioutil.ReadAll(tee)
			if err != nil {
				log.Error().Err(err).Msg("Error reading body")
				return false
			}
			predicate.results = ctx.Result{"payload": string(body)}
		case "json":
			var result interface{}
			d := json.NewDecoder(tee)
			if err := d.Decode(&result); err != nil {
				log.Debug().Err(err).Msg("Invalid JSON")
				return false
			}
			predicate.results = ctx.Result{"payload": result}
		}
		// Put back the body in case the predicate is used later
		c.In.Req.Body = ioutil.NopCloser(&b)
	}
	return true
}

// Generic interface //

// Result returns data set by the predicate
func (predicate *Predicate) Result() ctx.Result {
	return predicate.results
}

// Name returns the name of the predicate
func (predicate *Predicate) Name() string {
	return predicate.name
}

// Params returns a reference to a struct params accepted by the predicate
func (predicate *Predicate) Params() interface{} {
	return &predicate.params
}

// New returns a new Predicate
func New() genapid.Predicate {
	return &Predicate{
		name: Name,
	}
}
