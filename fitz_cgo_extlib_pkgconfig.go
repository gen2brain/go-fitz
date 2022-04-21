//go:build extlib && pkgconfig
// +build extlib,pkgconfig

package fitz

/*
#cgo pkg-config: mupdf
*/
import "C"
