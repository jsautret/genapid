// +build !disable_http

package plugins

import httppredicate "github.com/jsautret/go-api-broker/predicates/http"

func init() {
	Add(httppredicate.Name, httppredicate.New)
}
