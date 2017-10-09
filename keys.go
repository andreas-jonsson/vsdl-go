/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

package vsdl

import "fmt"

const (
	ReturnKey    Keycode = '\r'
	EscKey       Keycode = '\033'
	BackSpaceKey Keycode = '\b'
	TabKey       Keycode = '\t'
	SpaceKey     Keycode = ' '
	DeleteKey    Keycode = '\177'
)

const (
	NoMod         uint16 = 0x0000
	LeftShiftMod  uint16 = 0x0001
	RightShiftMod uint16 = 0x0002
	LeftCtrlMod   uint16 = 0x0040
	RightCtrlMod  uint16 = 0x0080
	LeftAltMod    uint16 = 0x0100
	RightAltMod   uint16 = 0x0200
	LeftGuiMod    uint16 = 0x0400
	RightGuiMod   uint16 = 0x0800
	NumMod        uint16 = 0x1000
	CapsMod       uint16 = 0x2000
	ModeMod       uint16 = 0x4000
)

const (
	CtrlMod  = LeftCtrlMod | RightCtrlMod
	ShiftMod = LeftShiftMod | RightShiftMod
	AltMod   = LeftAltMod | RightAltMod
	GuiMod   = LeftGuiMod | RightGuiMod
)

// Keycode (https://wiki.libsdl.org/SDL_Keycode)
type Keycode int32

// Keysym (https://wiki.libsdl.org/SDL_Keysym)
type Keysym struct {
	_   uint32
	Sym Keycode
	Mod uint16
	_   uint32
}

func (ks Keysym) String() string {
	return fmt.Sprintf("%c", ks.Sym)
}

func (ks Keysym) IsKey(k Keycode) bool {
	return ks.Sym == k
}

func (ks Keysym) IsMod(m uint16) bool {
	return ks.Mod&m != 0
}
