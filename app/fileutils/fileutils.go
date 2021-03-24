// Package fileutils provides file handling helpers
package fileutils

import (
	"os/user"
	"path/filepath"
	"strings"
)

var home string

func init() {
	usr, _ := user.Current()
	home = usr.HomeDir
}

// TODO automatically call Path using
// https://github.com/go-playground/mold

// Path returns a path, expanding ~ like the shell would do
func Path(path string) string {
	// Credit to https://stackoverflow.com/users/836390/joshlf
	//!\ Doesn't support ~user syntax
	if path == "~" {
		// In case of "~", which won't be caught by the "else if"
		return home

	} else if strings.HasPrefix(path, "~/") {
		// Use strings.HasPrefix so we don't match paths like
		// "/something/~/something/"
		return filepath.Join(home, path[2:])
	}
	return path
}
