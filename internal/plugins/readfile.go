// +build !disable_readfile

package plugins

import readfilepredicate "github.com/jsautret/go-api-broker/predicates/readfile"

func init() {
	Add(readfilepredicate.Name, readfilepredicate.New)
}
