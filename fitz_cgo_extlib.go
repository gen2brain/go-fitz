// +build extlib

package fitz

/*
#cgo LDFLAGS: -lmupdf -lmupdfthird -lm
#cgo android LDFLAGS: -llog
#cgo windows LDFLAGS: -lcomdlg32 -lgdi32
*/
import "C"
