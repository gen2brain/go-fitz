package fitz

// #include <mupdf/fitz.h>
// #cgo LDFLAGS: -lmupdf -lmujs -lopenjpeg -ljbig2dec -lz -lm -lfreetype -ljpeg -lpng -lbz2
// const char *fz_version = FZ_VERSION;
import "C"

import (
	"errors"
	"image"
	"os"
	"path/filepath"
	"unsafe"
)

// Fitz document
type Document struct {
	ctx *C.struct_fz_context_s
	doc *C.struct_fz_document_s
}

// NewDocument returns new fitz document
func NewDocument(filename string) (f *Document, err error) {
	f = &Document{}

	filename, err = filepath.Abs(filename)
	if err != nil {
		return
	}

	if _, e := os.Stat(filename); os.IsNotExist(e) {
		err = errors.New("fitz: no such file")
		return
	}

	f.ctx = (*C.struct_fz_context_s)(unsafe.Pointer(
		C.fz_new_context_imp(nil, nil, C.FZ_STORE_UNLIMITED, C.fz_version)))
	if f.ctx == nil {
		err = errors.New("fitz: cannot create context")
		return
	}

	C.fz_register_document_handlers(f.ctx)

	f.doc = C.fz_open_document(f.ctx, C.CString(filename))
	if f.doc == nil {
		err = errors.New("fitz: cannot open document")
		return
	}

	return
}

// Pages returns total number of pages in document
func (f *Document) Pages() int {
	return int(C.fz_count_pages(f.ctx, f.doc))
}

// Image returns image for given page number
func (f *Document) Image(page int) (image.Image, error) {
	var ctm C.fz_matrix
	C.fz_scale(&ctm, C.float(4.0), C.float(4.0))

	cs := C.fz_device_rgb(f.ctx)
	defer C.fz_drop_colorspace(f.ctx, cs)

	pixmap := C.fz_new_pixmap_from_page_number(f.ctx, f.doc, C.int(page), &ctm, cs)
	if pixmap == nil {
		return nil, errors.New("fitz: cannot create pixmap")
	}
	defer C.fz_drop_pixmap(f.ctx, pixmap)

	var bbox C.fz_irect
	C.fz_pixmap_bbox(f.ctx, pixmap, &bbox)

	pixels := C.fz_pixmap_samples(f.ctx, pixmap)
	if pixels == nil {
		return nil, errors.New("fitz: cannot get pixmap samples")
	}

	rect := image.Rect(int(bbox.x0), int(bbox.y0), int(bbox.x1), int(bbox.y1))
	bytes := C.GoBytes(unsafe.Pointer(pixels), 4*bbox.x1*bbox.y1)
	img := &image.RGBA{bytes, 4 * rect.Max.X, rect}

	return img, nil
}

// Close closes the underlying fitz document
func (f *Document) Close() {
	C.fz_drop_document(f.ctx, f.doc)
	C.fz_drop_context(f.ctx)
}
