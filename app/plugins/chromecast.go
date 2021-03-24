// +build !disable_chromecast

package plugins

import chromecastpredicate "github.com/jsautret/genapid/predicates/chromecast"

func init() {
	Add(chromecastpredicate.Name, chromecastpredicate.New)
}
