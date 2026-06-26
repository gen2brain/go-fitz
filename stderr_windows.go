//go:build (!cgo || nocgo) && windows

package fitz

import (
	"io"
	"os"
	"sync"
	"syscall"

	"github.com/ebitengine/purego"
	"golang.org/x/sys/windows"
)

const stderrFD = 2

// libmupdf writes through ucrtbase's fd table, not Go's os.Stderr, so the redirect goes through the CRT.
var (
	crtOnce          sync.Once
	crtDup           func(fd int32) int32
	crtDup2          func(fd1, fd2 int32) int32
	crtOpenOsfhandle func(handle uintptr, flags int32) int32
	crtClose         func(fd int32) int32
)

func crtInit() {
	h, err := syscall.LoadLibrary("ucrtbase.dll")
	if err != nil {
		return
	}

	purego.RegisterLibFunc(&crtDup, uintptr(h), "_dup")
	purego.RegisterLibFunc(&crtDup2, uintptr(h), "_dup2")
	purego.RegisterLibFunc(&crtOpenOsfhandle, uintptr(h), "_open_osfhandle")
	purego.RegisterLibFunc(&crtClose, uintptr(h), "_close")
}

func captureStderr(fn func()) string {
	crtOnce.Do(crtInit)
	if crtDup2 == nil {
		fn()
		return ""
	}

	saved := crtDup(stderrFD)
	if saved < 0 {
		fn()
		return ""
	}

	var hRead, hWrite windows.Handle
	if windows.CreatePipe(&hRead, &hWrite, nil, 0) != nil {
		crtClose(saved)
		fn()
		return ""
	}

	const oWronly = 0x0001
	fd := crtOpenOsfhandle(uintptr(hWrite), oWronly)
	if fd < 0 {
		windows.CloseHandle(hRead)
		windows.CloseHandle(hWrite)
		crtClose(saved)
		fn()
		return ""
	}

	crtDup2(fd, stderrFD)
	crtClose(fd)

	fn()

	crtDup2(saved, stderrFD)
	crtClose(saved)

	r := os.NewFile(uintptr(hRead), "stderr")
	buf, _ := io.ReadAll(r)
	r.Close()

	return string(buf)
}
