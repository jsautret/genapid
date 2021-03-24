// +build !disable_readfile

package plugins

import readfilepredicate "github.com/jsautret/genapid/predicates/readfile"

func init() {
	Add(readfilepredicate.Name, readfilepredicate.New)
}
