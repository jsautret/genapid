// +build !disable_body

package plugins

import bodypredicate "github.com/jsautret/genapid/predicates/body"

func init() {
	Add(bodypredicate.Name, bodypredicate.New)
}
