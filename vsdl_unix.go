// +build !windows

/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

package vsdl

/*
#include <SDL.h>
*/
import "C"
import (
	"errors"
	"image"
	"syscall"
	"unsafe"
)

const defaultLibName = ""

func loadLibrary(name string) (syscall.Handle, error) {
	return 0, nil
}

func sdlToGoError() error {
	if err := C.SDL_GetError(); err != nil {
		gostr := C.GoString(err)
		if len(gostr) > 0 {
			return errors.New(gostr)
		}
	}
	return nil
}

func sdlGetVersion() [3]byte {
	var version [3]byte
	C.SDL_GetVersion(unsafe.Pointer(&version[0]), unsafe.Pointer(&version[1]), unsafe.Pointer(&version[2]))
	return version
}

func sdlInit(flags uint32) bool {
	return C.SDL_Init(C.Uint32(flags)) != 0
}

func sdlCreateWindowAndRenderer(windowSize image.Point, windowPtr, rendererPtr uintptr) bool {
	ret := C.SDL_CreateWindowAndRenderer(
		C.int(windowSize.X),
		C.int(windowSize.Y),
		0,
		windowPtr,
		rendererPtr,
	)
	return ret != 0
}

func sdlDestroyRendererAndWindow(window, renderer uintptr) {
	C.SDL_DestroyRenderer(renderer)
	if err := sdlToGoError(); err != nil {
		log.Println(err)
	}
	C.SDL_DestroyWindow(window)
	if err := sdlToGoError(); err != nil {
		log.Println(err)
	}
}

func sdlRenderSetLogicalSize(renderer uintptr, logicalSize image.Point) bool {
	return C.SDL_RenderSetLogicalSize(renderer, C.int(logicalSize.X), C.int(logicalSize.Y)) != 0
}

func sdlCreateTexture(renderer uintptr, backBufferSize image.Point) uintptr {
	return C.SDL_CreateTexture(renderer, C.Uint32(pixelFormatABGR8888), C.int(1), C.int(backBufferSize.X), C.int(backBufferSize.Y))
}

func sdlDestroyTexture(texture uintptr) {
	C.SDL_DestroyTexture(texture)
}

func sdlUpdateTexture(texture, data, stride uintptr) bool {
	return C.SDL_UpdateTexture(texture, 0, data, stride) != 0
}

func sdlRenderCopy(renderer, texture uintptr) bool {
	return C.SDL_RenderCopy(renderer, texture, 0, 0) != 0
}

func sdlRenderPresent(renderer uintptr) {
	C.SDL_RenderPresent(renderer)
}

func sdlPollEvent(p uintptr) bool {
	return C.SDL_PollEvent(p) != 0
}
