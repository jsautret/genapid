// Copyright 2021 Jérôme Sautret. All rights reserved.  Use of this
// source code is governed by an Apache License 2.0 that can be found
// in the LICENSE file.

package httppredicate

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/jsautret/genapid/ctx"
	"github.com/jsautret/genapid/genapid"
	"github.com/rs/zerolog"
)

// Name of the predicate
var Name = "http"

// Predicate is a genapid.Predicate interface that describes the predicate
type Predicate struct {
	name   string
	params struct { // Params accepted by the predicate
		URL       string            `validate:"required,url"`
		Method    string            `validate:"required,oneof=GET HEAD OPTIONS POST PUT DELETE PATCH" mod:"default=GET,ucase"`
		Headers   map[string]string `validate:"dive,keys,required,endkeys" mapstructure:",omitempty"`
		Params    map[string]string `validate:"dive,keys,required,endkeys" mapstructure:",omitempty"`
		Body      *body             `mapstructure:",omitempty"`
		Response  string            `validate:"oneof=JSON STRING" mod:"default=STRING,ucase"`
		BasicAuth *basicAuth        `mapstructure:"basic_auth,omitempty"`
	}
	results ctx.Result // data returned by the http server
}

type body struct {
	JSON   interface{} `validate:"required_without_all=String,excluded_with=String"`
	String string      `validate:"required_without_all=JSON,excluded_with=JSON"`
}

type basicAuth struct{ Username, Password string }

// Call evaluates the predicate
func (predicate *Predicate) Call(log zerolog.Logger, c *ctx.Ctx) bool {
	p := predicate.params
	client := &http.Client{}
	var resp *http.Response
	var req *http.Request
	var err error

	log.Debug().Str("Method", p.Method).Msg("")

	if p.Params != nil {
		u, err := url.Parse(p.URL)
		if err != nil {
			log.Error().Err(err).Msg("Bad URL")
			return false
		}
		q := u.Query()
		for k, v := range p.Params {
			q.Set(k, v)
		}
		u.RawQuery = q.Encode()
		p.URL = u.String()
	}
	log.Debug().Str("URL", p.URL).Msg("")
	body, contentType := getBody(log, predicate)
	if body == nil {
		req, err = http.NewRequest(p.Method, p.URL, nil)
	} else {
		log.Debug().Interface("Body", body).Msg("")
		req, err = http.NewRequest(p.Method, p.URL, body)
	}
	if err != nil {
		log.Error().Err(err).Msg("")
		return false
	}
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	if p.Response == "JSON" {
		req.Header.Set("Accept", "application/json")
	} else {
		req.Header.Set("Accept", "*/*")
	}
	if p.BasicAuth != nil {
		req.Header.Set("Authorization", "Basic "+
			base64.StdEncoding.EncodeToString([]byte(
				p.BasicAuth.Username+":"+
					p.BasicAuth.Password)))
	}
	if p.Headers != nil {
		for k, v := range p.Headers {
			req.Header.Set(k, v)
		}
	}
	resp, err = client.Do(req)
	if err != nil {
		log.Error().Err(err).Msg("HTTP request failed")
		return false
	}
	predicate.results = ctx.Result{}
	switch p.Response {
	case "JSON":
		var result interface{}
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			log.Error().Err(err).Msg("Response is not JSON")
			return false
		}
		predicate.results["response"] = result
	default:
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Error().Err(errors.New("Cannot read response body"))
			return false
		}
		predicate.results["response"] = string(body)
	}
	predicate.results["type"] = resp.Header.Get("Content-Type")
	predicate.results["code"] = resp.StatusCode
	return true
}

func getBody(log zerolog.Logger, predicate *Predicate) (*bytes.Buffer, string) {
	p := predicate.params
	if p.Method == "POST" || p.Method == "PUT" || p.Method == "PATCH" {
		if p.Body == nil {
			log.Warn().Err(fmt.Errorf("Method %v should have a body",
				p.Method)).Msg("")
			return nil, ""
		}
		var postBody []byte
		var err error
		var contentType string
		if JSON := p.Body.JSON; JSON != nil {
			postBody, err = json.Marshal(JSON)
			if err != nil {
				log.Error().Err(err).Msg("body is not JSON")
				return nil, ""
			}
			contentType = "application/json"
		}
		if s := p.Body.String; s != "" {
			postBody = []byte(s)
			contentType = "text/plain"
		}
		return bytes.NewBuffer(postBody), contentType
	}
	return nil, ""
}

// Used for tests
func setHostURL(predicate genapid.Predicate, new string) error {
	p, _ := predicate.(*Predicate)
	URL, err := url.Parse(p.params.URL)
	if err != nil {
		return err
	}
	newURL, err := url.Parse(new)
	if err != nil {
		return err
	}
	URL.Host = newURL.Host
	URL.Scheme = newURL.Scheme
	p.params.URL = URL.String()
	return nil
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
