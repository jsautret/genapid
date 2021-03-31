// Copyright 2021 Jérôme Sautret. All rights reserved.  Use of this
// source code is governed by an Apache License 2.0 that can be found
// in the LICENSE file.

package utils

import "io"

// CloseQuietly closes `io.Closer` quietly. Very handy and helpful for code
// quality too.
func CloseQuietly(v interface{}) {
	if d, ok := v.(io.Closer); ok {
		_ = d.Close()
	}
}
