/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

package vsdl

import (
	"errors"
	"io/ioutil"
	"os"
	"path"
	"runtime"
	"syscall"
)

var (
	libraryHandle                syscall.Handle
	libraryName, removeDirectory string
)

var (
	sdlInitProc,
	sdlQuitProc,
	sdlGetErrorProc,
	sdlGetVersionProc,
	sdlCreateWindowAndRendererProc,
	sdlDestroyRendererProc,
	sdlDestroyWindowProc,
	sdlCreateTextureProc,
	sdlDestroyTextureProc,
	sdlUpdateTextureProc,
	sdlRenderCopyProc,
	sdlRenderPresentProc,
	sdlRenderSetLogicalSizeProc,
	sdlPollEventProc uintptr
)

func unloadLibrary() {
	syscall.FreeLibrary(libraryHandle)
	if removeDirectory != "" {
		os.RemoveAll(removeDirectory)
		removeDirectory = ""
	}

	libraryName = ""
	libraryHandle = 0
}

func loadEmbeddedLibrary(name string) (syscall.Handle, error) {
	if name != "" {
		return loadLibrary(name)
	}

	if sdlLibraryDataGOARCH != runtime.GOARCH {
		return 0, errors.New("this CPU architecture is not supported")
	}

	dir, err := ioutil.TempDir("", "")
	if err != nil {
		return 0, err
	}

	libraryName = path.Join(dir, defaultLibName)
	if err := ioutil.WriteFile(libraryName, sdlLibraryData, 0777); err != nil {
		return 0, err
	}
	removeDirectory = dir

	return loadLibrary(libraryName)
}

func getProc(name string) (uintptr, error) {
	proc, err := syscall.GetProcAddress(libraryHandle, name)
	if err != nil {
		return 0, newError(err, "could not get proc: "+name)
	}
	return proc, nil
}

func initProcs() error {
	var err error
	libraryHandle, err = loadEmbeddedLibrary(libraryName)
	if err != nil {
		return newError(err, "could not load library")
	}

	if sdlInitProc, err = getProc("SDL_Init"); err != nil {
		return err
	}

	if sdlQuitProc, err = getProc("SDL_Quit"); err != nil {
		return err
	}

	if sdlGetErrorProc, err = getProc("SDL_GetError"); err != nil {
		return err
	}

	if sdlGetVersionProc, err = getProc("SDL_GetVersion"); err != nil {
		return err
	}

	if sdlCreateWindowAndRendererProc, err = getProc("SDL_CreateWindowAndRenderer"); err != nil {
		return err
	}

	if sdlDestroyRendererProc, err = getProc("SDL_DestroyRenderer"); err != nil {
		return err
	}

	if sdlDestroyWindowProc, err = getProc("SDL_DestroyWindow"); err != nil {
		return err
	}

	if sdlCreateTextureProc, err = getProc("SDL_CreateTexture"); err != nil {
		return err
	}

	if sdlDestroyTextureProc, err = getProc("SDL_DestroyTexture"); err != nil {
		return err
	}

	if sdlUpdateTextureProc, err = getProc("SDL_UpdateTexture"); err != nil {
		return err
	}

	if sdlRenderCopyProc, err = getProc("SDL_RenderCopy"); err != nil {
		return err
	}

	if sdlRenderPresentProc, err = getProc("SDL_RenderPresent"); err != nil {
		return err
	}

	if sdlRenderSetLogicalSizeProc, err = getProc("SDL_RenderSetLogicalSize"); err != nil {
		return err
	}

	if sdlPollEventProc, err = getProc("SDL_PollEvent"); err != nil {
		return err
	}

	return nil
}
