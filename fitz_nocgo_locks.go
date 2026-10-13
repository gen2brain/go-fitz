//go:build !cgo || nocgo

package fitz

import (
	"sync"

	"github.com/ebitengine/purego"
)

// FZ_LOCK_ALLOC, FZ_LOCK_FREETYPE and FZ_LOCK_GLYPHCACHE, in that order.
// These mutexes must be shared by all MuPDF contexts in this process.
var contextMutexes [3]sync.Mutex

// Register callbacks once, since purego callbacks cannot be released.
// The unused return value is required by Windows callbacks.
var contextLocks = fzLocksContext{
	Lock: purego.NewCallback(func(_ purego.CDecl, _ uintptr, lock int32) uintptr {
		contextMutexes[lock].Lock()
		return 0
	}),
	Unlock: purego.NewCallback(func(_ purego.CDecl, _ uintptr, lock int32) uintptr {
		contextMutexes[lock].Unlock()
		return 0
	}),
}
