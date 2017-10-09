/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

package main

import (
	"image"
	"log"

	"github.com/andreas-jonsson/vsdl-go"
)

var (
	windowSize  = image.Pt(1280, 720)
	logicalSize = image.Pt(320, 180)
)

func main() {
	if err := vsdl.Initialize(vsdl.ConfigWithRenderer(windowSize, logicalSize)); err != nil {
		log.Fatalln(err)
	}
	defer vsdl.Shutdown()

	img := image.NewRGBA(image.Rectangle{Max: logicalSize})

	for {
		for ev := range vsdl.Events() {
			switch t := ev.(type) {
			case *vsdl.QuitEvent:
				return
			case *vsdl.KeyDownEvent:
				if t.Keysym.IsKey(vsdl.EscKey) {
					return
				}

				log.Printf("%s: %X\n", t.Keysym, t.Keysym.Sym)
			}
			ev.Release()
		}

		if err := vsdl.Present(img); err != nil {
			log.Println(err)
		}
	}
}
