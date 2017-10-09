// +build ignore

/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

package main

import (
	"archive/zip"
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"runtime"
)

func downloadAndExtract(url, name string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	br := bytes.NewReader(data)
	zr, err := zip.NewReader(br, br.Size())
	if err != nil {
		return err
	}

	for _, file := range zr.File {
		if file.Name == name {
			in, err := file.Open()
			if err != nil {
				return err
			}
			defer in.Close()

			out, err := os.Create(fmt.Sprintf("sdl_%s.go", runtime.GOOS))
			if err != nil {
				return err
			}
			defer out.Close()

			data, err := ioutil.ReadAll(in)
			if err != nil {
				return err
			}

			fmt.Fprintln(out, "// This file is generated.")
			fmt.Fprint(out, "\npackage vsdl\n\n")
			fmt.Fprintf(out, "const sdlLibraryDataGOARCH = %#v\n\n", runtime.GOARCH)
			fmt.Fprint(out, "var sdlLibraryData = ")
			fmt.Fprintf(out, "%#v\n", data)
			return nil
		}
	}

	return errors.New("could not locate library in archive")
}

func main() {
	if err := downloadAndExtract("https://www.libsdl.org/release/SDL2-2.0.6-win32-x64.zip", "SDL2.dll"); err != nil {
		panic(err)
	}
}
