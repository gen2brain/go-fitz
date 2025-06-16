//go:build (!cgo || nocgo) && unix && !darwin

package fitz

import (
	"fmt"
	"github.com/ebitengine/purego"
)

const (
	libname = "libmupdf.so"
)

// loadLibrary loads the so and panics on error.
func loadLibrary() uintptr {
	handle, err := purego.Dlopen(libname, purego.RTLD_NOW|purego.RTLD_GLOBAL)
	if err != nil {
		panic(fmt.Errorf("cannot load library: %w", err))
	}

	return handle
}

// procAddress returns the address of symbol name.
func procAddress(handle uintptr, procName string) uintptr {
	addr, err := purego.Dlsym(handle, procName)
	if err != nil {
		panic(fmt.Errorf("cannot get proc address for %s: %w", procName, err))
	}

	return addr
}
