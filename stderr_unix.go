//go:build (!cgo || nocgo) && unix

package fitz

import (
	"bytes"
	"io"
	"os"
	"syscall"

	"golang.org/x/sys/unix"
)

// captureStderr redirects fd 2 (where the MuPDF C runtime writes) into a buffer
// while fn runs, then restores it and returns what was written.
func captureStderr(fn func()) string {
	saved, err := unix.Dup(syscall.Stderr)
	if err != nil {
		fn()
		return ""
	}
	defer unix.Close(saved)

	r, w, err := os.Pipe()
	if err != nil {
		fn()
		return ""
	}

	if err := unix.Dup2(int(w.Fd()), syscall.Stderr); err != nil {
		w.Close()
		r.Close()
		fn()
		return ""
	}

	out := make(chan string, 1)
	go func() {
		var buf bytes.Buffer
		io.Copy(&buf, r)
		out <- buf.String()
	}()

	fn()

	unix.Dup2(saved, syscall.Stderr)
	w.Close()
	s := <-out
	r.Close()

	return s
}
