// +build extlib,!pkgconfig

package fitz

/*
#cgo !static LDFLAGS: -lmupdf -lm
#cgo static,!compat LDFLAGS: -lmupdf -lm -lmupdf-third
#cgo static,compat LDFLAGS: -lmupdf -lm -lmupdfthird
#cgo android LDFLAGS: -llog
#cgo windows LDFLAGS: -lcomdlg32 -lgdi32
*/
import "C"
