/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

package vsdl

import (
	"sync"
	"syscall"
	"unsafe"
)

func pollEvent() Event {
	ev := eventPool.Get().(*sdlEvent)
	up := unsafe.Pointer(ev)
	aev := (*anyEvent)(up)

	if ret, _, eno := syscall.Syscall(sdlPollEventProc, 1, uintptr(up), 0, 0); eno != 0 {
		panic(eno)
	} else if ret == 0 {
		aev.Release()
		return nil
	}

	switch *aev {
	case sdlQuitEventType:
		return (*QuitEvent)(up)
	case sdlWindowEventType:
		return (*WindowEvent)(up)
	case sdlKeyDownEventType:
		return (*KeyDownEvent)(up)
	case sdlKeyUpEventType:
		return (*KeyUpEvent)(up)
	case sdlMouseMotionEventType:
		return (*MouseMotionEvent)(up)
	case sdlMouseButtonDownEventType, sdlMouseButtonUpEventType:
		return (*MouseButtonEvent)(up)
	case sdlMouseWheelEventType:
		return (*MouseWheelEvent)(up)
	default:
		aev.Release()
		return nil
	}
}

const (
	sdlQuitEventType   = 0x100
	sdlWindowEventType = 0x200
)

const (
	sdlKeyDownEventType = 0x300 + iota
	sdlKeyUpEventType
)

const (
	sdlMouseMotionEventType = 0x400 + iota
	sdlMouseButtonDownEventType
	sdlMouseButtonUpEventType
	sdlMouseWheelEventType
)

const sdlEventMaxSize = 56

type sdlEvent [sdlEventMaxSize]byte

var eventPool = sync.Pool{
	New: func() interface{} {
		return new(sdlEvent)
	},
}

type Event interface {
	Release()
}

type anyEvent uint32

func (e *anyEvent) Release() {
	ev := (*sdlEvent)(unsafe.Pointer(e))
	eventPool.Put(ev)
}

type QuitEvent struct {
	anyEvent
}

// WindowEvent (https://wiki.libsdl.org/SDL_WindowEvent)
type WindowEvent struct {
	anyEvent

	_     uint32
	_     uint32
	Event uint8
	_     uint8
	_     uint8
	_     uint8
	Data1 int32
	Data2 int32
}

type KeyDownEvent struct {
	anyEvent

	_      uint32
	_      uint32
	State  uint8
	Repeat uint8
	_      uint8
	_      uint8
	Keysym Keysym
}

type KeyUpEvent struct {
	anyEvent

	_      uint32
	_      uint32
	State  uint8
	Repeat uint8
	_      uint8
	_      uint8
	Keysym Keysym
}

// MouseMotionEvent (https://wiki.libsdl.org/SDL_MouseMotionEvent)
type MouseMotionEvent struct {
	anyEvent

	_     uint32
	_     uint32
	Which uint32
	State uint32
	X     int32
	Y     int32
	XRel  int32
	YRel  int32
}

// MouseButtonEvent (https://wiki.libsdl.org/SDL_MouseButtonEvent)
type MouseButtonEvent struct {
	anyEvent

	_      uint32
	_      uint32
	Which  uint32
	Button uint8
	State  uint8
	_      uint8
	_      uint8
	X      int32
	Y      int32
}

// MouseWheelEvent (https://wiki.libsdl.org/SDL_MouseWheelEvent)
type MouseWheelEvent struct {
	anyEvent

	_         uint32
	_         uint32
	Which     uint32
	X         int32
	Y         int32
	Direction uint32
}
