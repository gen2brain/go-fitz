//go:build (!cgo || nocgo) && windows

package fitz

import (
	"fmt"
	"syscall"

	"golang.org/x/sys/windows"
)

const (
	libname = "libmupdf.dll"
)

// loadLibrary loads the dll and panics on error.
func loadLibrary() uintptr {
	handle, err := syscall.LoadLibrary(libname)
	if err != nil {
		panic(fmt.Errorf("cannot load library %s: %w", libname, err))
	}

	return uintptr(handle)
}

// procAddress returns the address of symbol name.
func procAddress(handle uintptr, procName string) uintptr {
	addr, err := windows.GetProcAddress(windows.Handle(handle), procName)
	if err != nil {
		panic(fmt.Errorf("cannot get proc address for %s: %w", procName, err))
	}

	return addr
}
