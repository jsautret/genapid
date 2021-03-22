// +build !disable_chromecast

package plugins

import chromecastpredicate "github.com/jsautret/go-api-broker/predicates/chromecast"

func init() {
	Add(chromecastpredicate.Name, chromecastpredicate.New)
}
