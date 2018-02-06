/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

//go:generate go run generate/main.go

package vsdl

import (
	"errors"
	"fmt"
	"image"
	"io/ioutil"
	logpkg "log"
	"runtime"
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

	version := sdlGetVersion()
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

		const sdlInitVideoFlag uint32 = 0x00000020

		if sdlInit(sdlInitVideoFlag) {
			errorChan <- sdlToGoError()
			return
		}
		defer sdlQuit()

		windowPtr := uintptr(unsafe.Pointer(&window))
		rendererPtr := uintptr(unsafe.Pointer(&renderer))

		if sdlCreateWindowAndRenderer(windowSize, windowPtr, rendererPtr) {
			errorChan <- sdlToGoError()
			return
		}
		defer sdlDestroyRendererAndWindow(window, renderer)

		if logicalSize.X != 0 {
			if sdlRenderSetLogicalSize(renderer, logicalSize) {
				errorChan <- sdlToGoError()
				return
			}
		}

		backBufferSize := windowSize
		if logicalSize.X != 0 {
			backBufferSize = logicalSize
		}

		if texture = sdlCreateTexture(renderer, backBufferSize); texture == 0 {
			errorChan <- sdlToGoError()
			return
		}
		defer sdlDestroyTexture(texture)

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
	eventChan := make(chan Event, maxEvents)
	sendCommand(true, func() error {
		for {
			ev := pollEvent()
			if ev == nil {
				close(eventChan)
				return nil
			}

			select {
			case eventChan <- ev:
			default:
				log.Println("event channel overflow")
			}
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
		if sdlUpdateTexture(texture, uintptr(unsafe.Pointer(&rgba.Pix[0])), uintptr(rgba.Stride)) {
			return sdlToGoError()
		}

		if sdlRenderCopy(renderer, texture) {
			return sdlToGoError()
		}

		sdlRenderPresent(renderer)
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
