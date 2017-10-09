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

func ConfigWithRenderer(size image.Point) Config {
	return func() error {
		windowSize = size
		return nil
	}
}

var log = logpkg.New(ioutil.Discard, "", logpkg.LstdFlags)

type command struct {
	f func() error
	a bool
}

var (
	errorChan   chan error
	commandChan chan command
)

var (
	windowSize                image.Point
	window, renderer, texture uintptr
)

func sdlToGoError() error {
	ret, _, eno := syscall.Syscall(sdlGetErrorProc, 0, 0, 0, 0)
	if eno != 0 {
		panic(eno)
	} else if ret == 0 {
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

	go func() {
		runtime.LockOSThread()
		defer runtime.UnlockOSThread()

		defer func() {
			unloadLibrary()
			errorChan <- nil
		}()

		var (
			ret uintptr
			eno syscall.Errno
		)

		const sdlInitVideoFlag uint32 = 0x00000020

		if ret, _, eno = syscall.Syscall(sdlInitProc, 1, uintptr(sdlInitVideoFlag), 0, 0); eno != 0 {
			log.Println(eno)
		}
		if ret != 0 {
			errorChan <- sdlToGoError()
			return
		}

		defer func() {
			if _, _, eno := syscall.Syscall(sdlQuitProc, 0, 0, 0, 0); eno != 0 {
				log.Println(eno)
			}
		}()

		windowPtr := uintptr(unsafe.Pointer(&window))
		rendererPtr := uintptr(unsafe.Pointer(&renderer))

		if ret, _, eno = syscall.Syscall6(sdlCreateWindowAndRendererProc, 5, uintptr(windowSize.X), uintptr(windowSize.Y), 0, windowPtr, rendererPtr, 0); eno != 0 {
			panic(eno)
		} else if ret != 0 {
			errorChan <- sdlToGoError()
			return
		}

		defer func() {
			if ret, _, eno = syscall.Syscall(sdlDestroyRendererProc, 1, renderer, 0, 0); eno != 0 || ret != 0 {
				log.Println("could not destroy renderer")
			}
			if ret, _, eno = syscall.Syscall(sdlDestroyWindowProc, 1, window, 0, 0); eno != 0 || ret != 0 {
				log.Println("could not destroy window")
			}
		}()

		if texture, _, eno = syscall.Syscall6(sdlCreateTextureProc, 5, renderer, 0, 1, uintptr(windowSize.X), uintptr(windowSize.Y), 0); eno != 0 {
			log.Println(eno)
		}
		if texture == 0 {
			errorChan <- sdlToGoError()
			return
		}

		defer func() {
			if ret, _, eno = syscall.Syscall(sdlDestroyTextureProc, 1, texture, 0, 0); eno != 0 || ret != 0 {
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
	if img.Bounds().Size() != windowSize {
		return errors.New("image is not the same size as the back-buffer")
	}

	rgba, ok := img.(*image.RGBA)
	if !ok {
		return errors.New("invalid image format")
	}

	return sendCommand(false, func() error {
		if ret, _, eno := syscall.Syscall6(sdlUpdateTextureProc, 4, texture, 0, uintptr(unsafe.Pointer(&rgba.Pix[0])), uintptr(rgba.Stride), 0, 0); eno != 0 {
			panic(eno)
		} else if ret != 0 {
			return sdlToGoError()
		}

		if ret, _, eno := syscall.Syscall6(sdlRenderCopyProc, 4, renderer, texture, 0, 0, 0, 0); eno != 0 {
			panic(eno)
		} else if ret != 0 {
			return sdlToGoError()
		}

		if _, _, eno := syscall.Syscall(sdlRenderPresentProc, 1, renderer, 0, 0); eno != 0 {
			panic(eno)
		}

		return nil
	})
}
