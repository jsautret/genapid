// +build !disable_header

package plugins

import headerpredicate "github.com/jsautret/genapid/predicates/header"

func init() {
	Add(headerpredicate.Name, headerpredicate.New)
}
