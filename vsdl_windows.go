/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

package vsdl

import (
	"syscall"
)

const defaultLibName = "SDL2.dll"

func loadLibrary(name string) (syscall.Handle, error) {
	dll, err := syscall.LoadDLL(name)
	if err != nil {
		return 0, err
	}
	return dll.Handle, err
}
