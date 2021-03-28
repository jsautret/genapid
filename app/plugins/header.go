// Copyright 2021 Jérôme Sautret. All rights reserved.  Use of this
// source code is governed by an Apache License 2.0 that can be found
// in the LICENSE file.

// +build !disable_header

package plugins

import headerpredicate "github.com/jsautret/genapid/predicates/header"

func init() {
	Add(headerpredicate.Name, headerpredicate.New)
}
