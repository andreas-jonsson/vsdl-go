/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

package main

import (
	"fmt"
	"image"

	"github.com/andreas-jonsson/vsdl"
)

func main() {
	if err := vsdl.Initialize(); err != nil {
		panic(err)
	}
	defer vsdl.Shutdown()

	img := image.NewRGBA(image.Rect(0, 0, 640, 480))

	for {
		for ev := range vsdl.Events() {
			switch t := ev.(type) {
			case *vsdl.QuitEvent:
				return
			case *vsdl.KeyDownEvent:
				if t.Keysym.IsKey(vsdl.EscKey) {
					return
				}

				fmt.Printf("%s: %X\n", t.Keysym, t.Keysym.Sym)
			}
			ev.Release()
		}

		if err := vsdl.Present(img); err != nil {
			panic(err)
		}
	}
}
