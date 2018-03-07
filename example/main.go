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
	f := func() error {
		img := image.NewRGBA(image.Rectangle{Max: logicalSize})

		for {
			for ev := range vsdl.Events() {
				switch t := ev.(type) {
				case *vsdl.QuitEvent:
					return nil
				case *vsdl.KeyDownEvent:
					if t.Keysym.IsKey(vsdl.EscKey) {
						return nil
					}

					log.Printf("%s: %X\n", t.Keysym, t.Keysym.Sym)
				}
				ev.Release()
			}

			wg, err := vsdl.Present(img)
			if err != nil {
				log.Println(err)
			}
			wg.Wait()
		}
	}

	if err := vsdl.Initialize(f, vsdl.ConfigWithRenderer(windowSize, logicalSize)); err != nil {
		log.Fatalln(err)
	}
}
