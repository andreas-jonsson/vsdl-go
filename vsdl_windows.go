/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

package vsdl

import (
	"C"
	"bytes"
	"errors"
	"image"
	"os"
	"syscall"
	"unsafe"
)

type libHandle = syscall.Handle

const defaultLibName = "SDL2.dll"

func loadLibrary(name string) (libHandle, error) {
	dll, err := syscall.LoadDLL(name)
	if err != nil {
		return 0, err
	}
	return dll.Handle, err
}

func unloadLibrary() {
	syscall.FreeLibrary(libraryHandle)
	if removeDirectory != "" {
		os.RemoveAll(removeDirectory)
		removeDirectory = ""
	}

	libraryName = ""
	libraryHandle = 0
}

func getProc(name string) (uintptr, error) {
	proc, err := syscall.GetProcAddress(libraryHandle, name)
	if err != nil {
		return 0, newError(err, "could not get proc: "+name)
	}
	return proc, nil
}

func sdlToGoError() error {
	ret, _, _ := syscall.Syscall(sdlGetErrorProc, 0, 0, 0, 0)
	if ret == 0 {
		return errors.New("unknown error")
	}

	var buf bytes.Buffer

	for {
		p := (*byte)(unsafe.Pointer(ret))
		if *p == 0 {
			return newError(errors.New(buf.String()), "internal error")
		}
		buf.WriteByte(*p)
	}
}

func sdlGetVersion() [3]byte {
	var version [3]byte
	syscall.Syscall(sdlGetVersionProc, 1, uintptr(unsafe.Pointer(&version)), 0, 0)
	return version
}

func sdlInit(flags uint32) bool {
	ret, _, _ := syscall.Syscall(sdlInitProc, 1, uintptr(flags), 0, 0)
	return ret != 0
}

func sdlQuit() {
	syscall.Syscall(sdlShowCursorProc, 1, 1, 0, 0)
	syscall.Syscall(sdlQuitProc, 0, 0, 0, 0)
}

func sdlCreateWindowAndRenderer(windowSize image.Point, windowPtr, rendererPtr uintptr) bool {
	ret, _, _ := syscall.Syscall6(sdlCreateWindowAndRendererProc, 5, uintptr(windowSize.X), uintptr(windowSize.Y), 0, windowPtr, rendererPtr, 0)
	if ret == 0 {
		syscall.Syscall(sdlShowCursorProc, 1, 0, 0, 0)
		return false
	}
	return true
}

func sdlDestroyRendererAndWindow(window, renderer uintptr) {
	if ret, _, _ := syscall.Syscall(sdlDestroyRendererProc, 1, renderer, 0, 0); ret != 0 {
		log.Println("could not destroy renderer")
	}
	if ret, _, _ := syscall.Syscall(sdlDestroyWindowProc, 1, window, 0, 0); ret != 0 {
		log.Println("could not destroy window")
	}
}

func sdlRenderSetLogicalSize(renderer uintptr, logicalSize image.Point) bool {
	ret, _, _ := syscall.Syscall(sdlRenderSetLogicalSizeProc, 3, renderer, uintptr(logicalSize.X), uintptr(logicalSize.Y))
	return ret != 0
}

func sdlToggleFullscreen(window uintptr) (bool, error) {
	flags, _, _ := syscall.Syscall(sdlGetWindowFlagsProc, 1, window, 0, 0)
	isFullscreen := (uint32(flags) & sdl_WINDOW_FULLSCREEN) != 0

	if isFullscreen {
		syscall.Syscall(sdlSetWindowFullscreenProc, 2, window, 0, 0)
		return false, sdlToGoError()
	}

	syscall.Syscall(sdlSetWindowFullscreenProc, 2, window, uintptr(defaultFullscreenFlag), uintptr(0))
	return true, sdlToGoError()
}

func sdlCreateTexture(renderer uintptr, backBufferSize image.Point) uintptr {
	texture, _, _ := syscall.Syscall6(sdlCreateTextureProc, 5, renderer, uintptr(pixelFormatABGR8888), 1, uintptr(backBufferSize.X), uintptr(backBufferSize.Y), 0)
	return texture
}

func sdlDestroyTexture(texture uintptr) {
	if ret, _, _ := syscall.Syscall(sdlDestroyTextureProc, 1, texture, 0, 0); ret != 0 {
		log.Println("could not destroy texture")
	}
}

func sdlUpdateTexture(texture, data, stride uintptr) bool {
	ret, _, _ := syscall.Syscall6(sdlUpdateTextureProc, 4, texture, 0, data, stride, 0, 0)
	return ret != 0
}

func sdlRenderCopy(renderer, texture uintptr) bool {
	ret, _, _ := syscall.Syscall6(sdlRenderCopyProc, 4, renderer, texture, 0, 0, 0, 0)
	return ret != 0
}

func sdlRenderPresent(renderer uintptr) {
	syscall.Syscall(sdlRenderPresentProc, 1, renderer, 0, 0)
}

func sdlPollEvent(p uintptr) bool {
	ret, _, _ := syscall.Syscall(sdlPollEventProc, 1, p, 0, 0)
	return ret != 0
}
