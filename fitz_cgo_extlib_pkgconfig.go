//go:build cgo && !nocgo && extlib && pkgconfig

package fitz

/*
#cgo !static pkg-config: mupdf
#cgo static pkg-config: --static mupdf
*/
import "C"
