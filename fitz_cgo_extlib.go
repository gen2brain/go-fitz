//go:build extlib && !pkgconfig

package fitz

/*
#cgo !static LDFLAGS: -lmupdf -lm
#cgo static LDFLAGS: -lmupdf -lm -lmupdf-third
#cgo android LDFLAGS: -llog
#cgo windows LDFLAGS: -lcomdlg32 -lgdi32
*/
import "C"
