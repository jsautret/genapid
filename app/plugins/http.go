// +build !disable_http

package plugins

import httppredicate "github.com/jsautret/genapid/predicates/http"

func init() {
	Add(httppredicate.Name, httppredicate.New)
}
