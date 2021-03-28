// Copyright 2021 Jérôme Sautret. All rights reserved.  Use of this
// source code is governed by an Apache License 2.0 that can be found
// in the LICENSE file.

// +build !disable_chromecast

package plugins

import chromecastpredicate "github.com/jsautret/genapid/predicates/chromecast"

func init() {
	Add(chromecastpredicate.Name, chromecastpredicate.New)
}
