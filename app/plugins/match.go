// Copyright 2021 Jérôme Sautret. All rights reserved.  Use of this
// source code is governed by an Apache License 2.0 that can be found
// in the LICENSE file.

// +build !disable_match

package plugins

import matchpredicate "github.com/jsautret/genapid/predicates/match"

func init() {
	Add(matchpredicate.Name, matchpredicate.New)
}
