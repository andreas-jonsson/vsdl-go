// +build !windows

/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

package vsdl

/*
#cgo linux freebsd darwin pkg-config: sdl2
#include <SDL.h>
*/
import "C"
import (
	"errors"
	"image"
	"unsafe"
)

type libHandle = uintptr

const defaultLibName = ""

func loadLibrary(name string) (libHandle, error) {
	return 0, nil
}

func unloadLibrary() {
	libraryName = ""
	libraryHandle = 0
}

func getProc(name string) (uintptr, error) {
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
	C.SDL_GetVersion((*C.SDL_version)(unsafe.Pointer(&version[0])))
	return version
}

func sdlInit(flags uint32) bool {
	return C.SDL_Init(C.Uint32(flags)) != 0
}

func sdlQuit() {
	C.SDL_Quit()
}

func sdlCreateWindowAndRenderer(windowSize image.Point, windowPtr, rendererPtr uintptr) bool {
	ret := C.SDL_CreateWindowAndRenderer(
		C.int(windowSize.X),
		C.int(windowSize.Y),
		0,
		(**C.SDL_Window)(unsafe.Pointer(windowPtr)),
		(**C.SDL_Renderer)(unsafe.Pointer(rendererPtr)),
	)
	return ret != 0
}

func sdlDestroyRendererAndWindow(window, renderer uintptr) {
	C.SDL_DestroyRenderer((*C.SDL_Renderer)(unsafe.Pointer(renderer)))
	if err := sdlToGoError(); err != nil {
		log.Println(err)
	}
	C.SDL_DestroyWindow((*C.SDL_Window)(unsafe.Pointer(window)))
	if err := sdlToGoError(); err != nil {
		log.Println(err)
	}
}

func sdlRenderSetLogicalSize(renderer uintptr, logicalSize image.Point) bool {
	return C.SDL_RenderSetLogicalSize((*C.SDL_Renderer)(unsafe.Pointer(renderer)), C.int(logicalSize.X), C.int(logicalSize.Y)) != 0
}

func sdlCreateTexture(renderer uintptr, backBufferSize image.Point) uintptr {
	return uintptr(unsafe.Pointer(C.SDL_CreateTexture((*C.SDL_Renderer)(unsafe.Pointer(renderer)), C.Uint32(pixelFormatABGR8888), C.int(1), C.int(backBufferSize.X), C.int(backBufferSize.Y))))
}

func sdlDestroyTexture(texture uintptr) {
	C.SDL_DestroyTexture((*C.SDL_Texture)(unsafe.Pointer(texture)))
}

func sdlUpdateTexture(texture, data, stride uintptr) bool {
	return C.SDL_UpdateTexture((*C.SDL_Texture)(unsafe.Pointer(texture)), nil, unsafe.Pointer(data), C.int(stride)) != 0
}

func sdlRenderCopy(renderer, texture uintptr) bool {
	return C.SDL_RenderCopy((*C.SDL_Renderer)(unsafe.Pointer(renderer)), (*C.SDL_Texture)(unsafe.Pointer(texture)), nil, nil) != 0
}

func sdlRenderPresent(renderer uintptr) {
	C.SDL_RenderPresent((*C.SDL_Renderer)(unsafe.Pointer(renderer)))
}

func sdlPollEvent(p uintptr) bool {
	return C.SDL_PollEvent((*C.SDL_Event)(unsafe.Pointer(p))) != 0
}
