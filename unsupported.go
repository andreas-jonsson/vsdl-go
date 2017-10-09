// +build !windows

/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

package vsdl

import (
	"syscall"
)

const defaultLibName = ""

var sdlLibraryData []byte

func init() {
	panic("this operating-system is not supported")
}

func loadLibrary(name string) (syscall.Handle, error) {
	return 0, nil
}
