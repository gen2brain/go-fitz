// +build extlib

package fitz

/*
#cgo !static LDFLAGS: -lmupdf -lm
#cgo static LDFLAGS: -lmupdf -lm -lmupdfthird
#cgo android LDFLAGS: -llog
#cgo windows LDFLAGS: -lcomdlg32 -lgdi32
*/
import "C"
