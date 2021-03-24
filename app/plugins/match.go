// +build !disable_match

package plugins

import matchpredicate "github.com/jsautret/genapid/predicates/match"

func init() {
	Add(matchpredicate.Name, matchpredicate.New)
}
