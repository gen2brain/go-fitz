//go:build (!cgo || nocgo) && darwin

package fitz

import (
	"fmt"

	"github.com/ebitengine/purego"
)

const (
	libname = "libmupdf.dylib"
)

// loadLibrary loads the so and panics on error.
func loadLibrary() uintptr {
	handle, err := purego.Dlopen(libname, purego.RTLD_NOW|purego.RTLD_GLOBAL)
	if err != nil {
		panic(fmt.Errorf("cannot load library: %w", err))
	}

	return uintptr(handle)
}
