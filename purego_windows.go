//go:build (!cgo || nocgo) && windows

package fitz

import (
	"fmt"
	"syscall"
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
