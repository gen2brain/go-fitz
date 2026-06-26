// Package fitz provides wrapper for the [MuPDF](http://mupdf.com/) fitz library
// that can extract pages from PDF, EPUB, MOBI, DOCX, XLSX and PPTX documents as IMG, TXT, HTML or SVG.
package fitz

import (
	"errors"
	"image"
	"math"
	"unsafe"
)

// Errors.
var (
	ErrNoSuchFile      = errors.New("fitz: no such file")
	ErrCreateContext   = errors.New("fitz: cannot create context")
	ErrOpenDocument    = errors.New("fitz: cannot open document")
	ErrEmptyBytes      = errors.New("fitz: cannot send empty bytes")
	ErrOpenMemory      = errors.New("fitz: cannot open memory")
	ErrLoadPage        = errors.New("fitz: cannot load page")
	ErrRunPageContents = errors.New("fitz: cannot run page contents")
	ErrPageMissing     = errors.New("fitz: page missing")
	ErrCreatePixmap    = errors.New("fitz: cannot create pixmap")
	ErrPixmapSamples   = errors.New("fitz: cannot get pixmap samples")
	ErrNeedsPassword   = errors.New("fitz: document needs password")
	ErrLoadOutline     = errors.New("fitz: cannot load outline")
)

// MaxStore is maximum size in bytes of the resource store, before it will start evicting cached resources such as fonts and images.
var MaxStore = 256 << 20

// FzVersion is used for experimental purego implementation, it must be exactly the same as libmupdf shared library version.
// It is also possible to set `FZ_VERSION` environment variable.
var FzVersion = "1.28.0"

// Outline type.
type Outline struct {
	// Hierarchy level of the entry (starting from 1).
	Level int
	// Title of outline item.
	Title string
	// Destination in the document to be displayed when this outline item is activated.
	URI string
	// The page number of an internal link.
	Page int
	// Top.
	Top float64
}

// Link type.
type Link struct {
	URI string
}

func bytePtrToString(p *byte) string {
	if p == nil {
		return ""
	}
	if *p == 0 {
		return ""
	}

	// Find NUL terminator.
	n := 0
	for ptr := unsafe.Pointer(p); *(*byte)(ptr) != 0; n++ {
		ptr = unsafe.Pointer(uintptr(ptr) + 1)
	}

	return string(unsafe.Slice(p, n))
}

// cropImage returns the img region at point rectangle (x0,y0)-(x1,y1) scaled to dpi, at origin (0,0).
func cropImage(img *image.RGBA, x0, y0, x1, y1, dpi float64) *image.RGBA {
	s := dpi / 72

	r := image.Rect(int(math.Round(x0*s)), int(math.Round(y0*s)), int(math.Round(x1*s)), int(math.Round(y1*s)))
	r = r.Intersect(img.Bounds())
	if r.Empty() {
		return img
	}

	out := image.NewRGBA(image.Rect(0, 0, r.Dx(), r.Dy()))
	for y := 0; y < r.Dy(); y++ {
		so := img.PixOffset(r.Min.X, r.Min.Y+y)
		do := out.PixOffset(0, y)
		copy(out.Pix[do:do+r.Dx()*4], img.Pix[so:so+r.Dx()*4])
	}

	return out
}
