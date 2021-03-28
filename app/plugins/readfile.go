// Copyright 2021 Jérôme Sautret. All rights reserved.  Use of this
// source code is governed by an Apache License 2.0 that can be found
// in the LICENSE file.

// +build !disable_readfile

package plugins

import readfilepredicate "github.com/jsautret/genapid/predicates/readfile"

func init() {
	Add(readfilepredicate.Name, readfilepredicate.New)
}
