// +build extlib

package fitz

/*
#cgo LDFLAGS: -lmupdf -lm
#cgo android LDFLAGS: -llog
#cgo windows LDFLAGS: -lcomdlg32 -lgdi32
*/
import "C"
