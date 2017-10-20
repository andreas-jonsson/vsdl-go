/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

//go:generate go run generate/main.go

package vsdl

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"io/ioutil"
	logpkg "log"
	"runtime"
	"syscall"
	"unsafe"
)

type Error struct {
	String   string
	Internal error
}

func (e *Error) Error() string {
	return e.String
}

func newError(internal error, format string, args ...interface{}) *Error {
	return &Error{
		Internal: internal,
		String:   fmt.Sprintf(format, args...),
	}
}

type Config func() error

func ConfigWithLibrary(p string) Config {
	return func() error {
		libraryName = p
		return nil
	}
}

func ConfigWithLogger(l *logpkg.Logger) Config {
	return func() error {
		log = l
		return nil
	}
}

func ConfigWithRenderer(size, logical image.Point) Config {
	return func() error {
		windowSize = size
		logicalSize = logical
		return nil
	}
}

var log = logpkg.New(ioutil.Discard, "", logpkg.LstdFlags)

var sdlExpectedVersion = [2]byte{2, 0}

type command struct {
	f func() error
	a bool
}

var (
	errorChan   chan error
	commandChan chan command
)

var (
	windowSize, logicalSize   image.Point
	window, renderer, texture uintptr
)

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

func sendCommand(async bool, f func() error) error {
	commandChan <- command{f, async}
	if !async {
		return <-errorChan
	}
	return nil
}

func Initialize(configs ...Config) error {
	windowSize = image.Point{640, 480}
	logicalSize = image.Point{}
	errorChan = make(chan error)
	commandChan = make(chan command)

	for _, cfg := range configs {
		if err := cfg(); err != nil {
			return newError(err, "configuration error")
		}
	}

	if err := initProcs(); err != nil {
		return nil
	}

	var version [3]byte
	syscall.Syscall(sdlGetVersionProc, 1, uintptr(unsafe.Pointer(&version)), 0, 0)

	if version[0] != sdlExpectedVersion[0] || version[1] != sdlExpectedVersion[1] {
		log.Printf("Expected SDL version %d.%d.x, but version %d.%d.%d was loaded.\n", sdlExpectedVersion[0], sdlExpectedVersion[1], version[0], version[1], version[2])
	}

	go func() {
		runtime.LockOSThread()
		defer runtime.UnlockOSThread()

		defer func() {
			unloadLibrary()
			errorChan <- nil
		}()

		var ret uintptr

		const sdlInitVideoFlag uint32 = 0x00000020

		if ret, _, _ = syscall.Syscall(sdlInitProc, 1, uintptr(sdlInitVideoFlag), 0, 0); ret != 0 {
			errorChan <- sdlToGoError()
			return
		}

		defer func() {
			syscall.Syscall(sdlQuitProc, 0, 0, 0, 0)
		}()

		windowPtr := uintptr(unsafe.Pointer(&window))
		rendererPtr := uintptr(unsafe.Pointer(&renderer))

		if ret, _, _ = syscall.Syscall6(sdlCreateWindowAndRendererProc, 5, uintptr(windowSize.X), uintptr(windowSize.Y), 0, windowPtr, rendererPtr, 0); ret != 0 {
			errorChan <- sdlToGoError()
			return
		}

		defer func() {
			if ret, _, _ = syscall.Syscall(sdlDestroyRendererProc, 1, renderer, 0, 0); ret != 0 {
				log.Println("could not destroy renderer")
			}
			if ret, _, _ = syscall.Syscall(sdlDestroyWindowProc, 1, window, 0, 0); ret != 0 {
				log.Println("could not destroy window")
			}
		}()

		if logicalSize.X != 0 {
			if ret, _, _ = syscall.Syscall(sdlRenderSetLogicalSizeProc, 3, renderer, uintptr(logicalSize.X), uintptr(logicalSize.Y)); ret != 0 {
				errorChan <- sdlToGoError()
				return
			}
		}

		backBufferSize := windowSize
		if logicalSize.X != 0 {
			backBufferSize = logicalSize
		}

		if texture, _, _ = syscall.Syscall6(sdlCreateTextureProc, 5, renderer, uintptr(pixelFormatABGR8888), 1, uintptr(backBufferSize.X), uintptr(backBufferSize.Y), 0); texture == 0 {
			errorChan <- sdlToGoError()
			return
		}

		defer func() {
			if ret, _, _ = syscall.Syscall(sdlDestroyTextureProc, 1, texture, 0, 0); ret != 0 {
				log.Println("could not destroy texture")
			}
		}()

		errorChan <- nil

		for c := range commandChan {
			err := c.f()
			if !c.a {
				errorChan <- err
			}
		}
	}()

	return <-errorChan
}

func Shutdown() error {
	sendCommand(true, func() error {
		close(commandChan)
		return nil
	})
	return <-errorChan
}

func Events() <-chan Event {
	eventChan := make(chan Event)
	sendCommand(true, func() error {
		for {
			ev := pollEvent()
			if ev == nil {
				close(eventChan)
				return nil
			}
			eventChan <- ev
		}
	})
	return eventChan
}

func Present(img image.Image) error {
	imgSize := img.Bounds().Size()
	backBufferSize := windowSize

	if logicalSize.X != 0 {
		backBufferSize = logicalSize
	}

	if imgSize != backBufferSize {
		return errors.New("image is not the same size as the back-buffer")
	}

	rgba, ok := img.(*image.RGBA)
	if !ok {
		return errors.New("invalid image format")
	}

	return sendCommand(false, func() error {
		if ret, _, _ := syscall.Syscall6(sdlUpdateTextureProc, 4, texture, 0, uintptr(unsafe.Pointer(&rgba.Pix[0])), uintptr(rgba.Stride), 0, 0); ret != 0 {
			return sdlToGoError()
		}

		if ret, _, _ := syscall.Syscall6(sdlRenderCopyProc, 4, renderer, texture, 0, 0, 0, 0); ret != 0 {
			return sdlToGoError()
		}

		syscall.Syscall(sdlRenderPresentProc, 1, renderer, 0, 0)
		return nil
	})
}

func definePixelFormat(ty, order, layout, bits, bytes uint32) uint32 {
	return (1 << 28) | (ty << 24) | (order << 20) | (layout << 16) | (bits << 8) | (bytes << 0)
}

const (
	pixelTypePacked32 = 6
	packedOrderABGR   = 7
	packedLayout8888  = 6
)

var pixelFormatABGR8888 = definePixelFormat(pixelTypePacked32, packedOrderABGR, packedLayout8888, 32, 4)
