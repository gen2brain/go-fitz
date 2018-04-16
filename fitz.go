// Package fitz provides wrapper for the [MuPDF](http://mupdf.com/) fitz library
// that can extract pages from PDF, EPUB and XPS documents as images or text.
package fitz

/*
#include <mupdf/fitz.h>
#include <stdlib.h>

const char *fz_version = FZ_VERSION;
*/
import "C"

import (
	"errors"
	"image"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"unsafe"
)

// Errors.
var (
	ErrNoSuchFile    = errors.New("fitz: no such file")
	ErrCreateContext = errors.New("fitz: cannot create context")
	ErrOpenDocument  = errors.New("fitz: cannot open document")
	ErrOpenMemory    = errors.New("fitz: cannot open memory")
	ErrPageMissing   = errors.New("fitz: page missing")
	ErrCreatePixmap  = errors.New("fitz: cannot create pixmap")
	ErrPixmapSamples = errors.New("fitz: cannot get pixmap samples")
	ErrNeedsPassword = errors.New("fitz: document needs password")
)

// Document represents fitz document.
type Document struct {
	ctx *C.struct_fz_context_s
	doc *C.struct_fz_document_s
}

// New returns new fitz document.
func New(filename string) (f *Document, err error) {
	f = &Document{}

	filename, err = filepath.Abs(filename)
	if err != nil {
		return
	}

	if _, e := os.Stat(filename); e != nil {
		err = ErrNoSuchFile
		return
	}

	f.ctx = (*C.struct_fz_context_s)(unsafe.Pointer(C.fz_new_context_imp(nil, nil, C.FZ_STORE_UNLIMITED, C.fz_version)))
	if f.ctx == nil {
		err = ErrCreateContext
		return
	}

	C.fz_register_document_handlers(f.ctx)

	cfilename := C.CString(filename)
	defer C.free(unsafe.Pointer(cfilename))

	f.doc = C.fz_open_document(f.ctx, cfilename)
	if f.doc == nil {
		err = ErrOpenDocument
	}

	ret := C.fz_needs_password(f.ctx, f.doc)
	v := bool(int(ret) != 0)
	if v {
		err = ErrNeedsPassword
	}

	return
}

// NewFromMemory returns new fitz document from byte slice.
func NewFromMemory(b []byte) (f *Document, err error) {
	f = &Document{}

	f.ctx = (*C.struct_fz_context_s)(unsafe.Pointer(C.fz_new_context_imp(nil, nil, C.FZ_STORE_UNLIMITED, C.fz_version)))
	if f.ctx == nil {
		err = ErrCreateContext
		return
	}

	C.fz_register_document_handlers(f.ctx)

	data := (*C.uchar)(C.CBytes(b))

	stream := C.fz_open_memory(f.ctx, data, C.size_t(len(b)))
	if stream == nil {
		err = ErrOpenMemory
		return
	}

	cmagic := C.CString("application/pdf")
	defer C.free(unsafe.Pointer(cmagic))

	f.doc = C.fz_open_document_with_stream(f.ctx, cmagic, stream)
	if f.doc == nil {
		err = ErrOpenDocument
	}

	ret := C.fz_needs_password(f.ctx, f.doc)
	v := bool(int(ret) != 0)
	if v {
		err = ErrNeedsPassword
	}

	return
}

// NewFromReader returns new fitz document from io.Reader.
func NewFromReader(r io.Reader) (f *Document, err error) {
	b, e := ioutil.ReadAll(r)
	if e != nil {
		err = e
		return
	}

	f, err = NewFromMemory(b)

	return
}

// NumPage returns total number of pages in document.
func (f *Document) NumPage() int {
	return int(C.fz_count_pages(f.ctx, f.doc))
}

// Image returns image for given page number.
func (f *Document) Image(pageNumber int) (image.Image, error) {
	if pageNumber >= f.NumPage() {
		return nil, ErrPageMissing
	}

	page := C.fz_load_page(f.ctx, f.doc, C.int(pageNumber))
	defer C.fz_drop_page(f.ctx, page)

	var bounds C.fz_rect
	C.fz_bound_page(f.ctx, page, &bounds)

	var ctm C.fz_matrix
	C.fz_scale(&ctm, C.float(300.0/72), C.float(300.0/72))

	var bbox C.fz_irect
	C.fz_transform_rect(&bounds, &ctm)
	C.fz_round_rect(&bbox, &bounds)

	pixmap := C.fz_new_pixmap_with_bbox(f.ctx, C.fz_device_rgb(f.ctx), &bbox, nil, 1)
	if pixmap == nil {
		return nil, ErrCreatePixmap
	}

	C.fz_clear_pixmap_with_value(f.ctx, pixmap, C.int(0xff))
	defer C.fz_drop_pixmap(f.ctx, pixmap)

	device := C.fz_new_draw_device(f.ctx, &ctm, pixmap)
	defer C.fz_drop_device(f.ctx, device)

	drawMatrix := C.fz_identity
	C.fz_run_page(f.ctx, page, device, &drawMatrix, nil)

	C.fz_close_device(f.ctx, device)

	pixels := C.fz_pixmap_samples(f.ctx, pixmap)
	if pixels == nil {
		return nil, ErrPixmapSamples
	}

	rect := image.Rect(int(bbox.x0), int(bbox.y0), int(bbox.x1), int(bbox.y1))
	bytes := C.GoBytes(unsafe.Pointer(pixels), C.int(4*bbox.x1*bbox.y1))
	img := &image.RGBA{Pix: bytes, Stride: 4 * rect.Max.X, Rect: rect}

	return img, nil
}

// Text returns text for given page number.
func (f *Document) Text(pageNumber int) (string, error) {
	if pageNumber >= f.NumPage() {
		return "", ErrPageMissing
	}

	page := C.fz_load_page(f.ctx, f.doc, C.int(pageNumber))
	defer C.fz_drop_page(f.ctx, page)

	var bounds C.fz_rect
	C.fz_bound_page(f.ctx, page, &bounds)

	var ctm C.fz_matrix
	C.fz_scale(&ctm, C.float(72.0/72), C.float(72.0/72))

	text := C.fz_new_stext_page(f.ctx, &bounds)
	defer C.fz_drop_stext_page(f.ctx, text)

	var opts C.fz_stext_options
	opts.flags = 0

	device := C.fz_new_stext_device(f.ctx, text, &opts)
	defer C.fz_drop_device(f.ctx, device)

	var cookie C.fz_cookie
	C.fz_run_page(f.ctx, page, device, &ctm, &cookie)

	C.fz_close_device(f.ctx, device)

	buf := C.fz_new_buffer_from_stext_page(f.ctx, text)
	defer C.fz_drop_buffer(f.ctx, buf)

	out := C.fz_new_output_with_buffer(f.ctx, buf)
	defer C.fz_drop_output(f.ctx, out)

	C.fz_print_stext_page_as_text(f.ctx, out, text)
	str := C.GoString(C.fz_string_from_buffer(f.ctx, buf))

	return str, nil
}

// HTML returns html for given page number.
func (f *Document) HTML(pageNumber int, header bool) (string, error) {
	if pageNumber >= f.NumPage() {
		return "", ErrPageMissing
	}

	page := C.fz_load_page(f.ctx, f.doc, C.int(pageNumber))
	defer C.fz_drop_page(f.ctx, page)

	var bounds C.fz_rect
	C.fz_bound_page(f.ctx, page, &bounds)

	var ctm C.fz_matrix
	C.fz_scale(&ctm, C.float(72.0/72), C.float(72.0/72))

	text := C.fz_new_stext_page(f.ctx, &bounds)
	defer C.fz_drop_stext_page(f.ctx, text)

	var opts C.fz_stext_options
	opts.flags = C.FZ_STEXT_PRESERVE_IMAGES

	device := C.fz_new_stext_device(f.ctx, text, &opts)
	defer C.fz_drop_device(f.ctx, device)

	var cookie C.fz_cookie
	C.fz_run_page(f.ctx, page, device, &ctm, &cookie)

	C.fz_close_device(f.ctx, device)

	buf := C.fz_new_buffer(f.ctx, 1024)
	defer C.fz_drop_buffer(f.ctx, buf)

	out := C.fz_new_output_with_buffer(f.ctx, buf)
	defer C.fz_drop_output(f.ctx, out)

	if header {
		C.fz_print_stext_header_as_html(f.ctx, out)
	}
	C.fz_print_stext_page_as_html(f.ctx, out, text)
	if header {
		C.fz_print_stext_trailer_as_html(f.ctx, out)
	}

	str := C.GoString(C.fz_string_from_buffer(f.ctx, buf))

	return str, nil
}

// SVG returns svg document for given page number.
func (f *Document) SVG(pageNumber int) (string, error) {
	if pageNumber >= f.NumPage() {
		return "", ErrPageMissing
	}

	page := C.fz_load_page(f.ctx, f.doc, C.int(pageNumber))
	defer C.fz_drop_page(f.ctx, page)

	var bounds C.fz_rect
	C.fz_bound_page(f.ctx, page, &bounds)

	var ctm C.fz_matrix
	C.fz_scale(&ctm, C.float(72.0/72), C.float(72.0/72))
	C.fz_transform_rect(&bounds, &ctm)

	buf := C.fz_new_buffer(f.ctx, 1024)
	defer C.fz_drop_buffer(f.ctx, buf)

	out := C.fz_new_output_with_buffer(f.ctx, buf)
	defer C.fz_drop_output(f.ctx, out)

	device := C.fz_new_svg_device(f.ctx, out, bounds.x1-bounds.x0, bounds.y1-bounds.y0, C.FZ_SVG_TEXT_AS_PATH, 1)
	//defer C.fz_drop_device(f.ctx, device)

	var cookie C.fz_cookie
	C.fz_run_page(f.ctx, page, device, &ctm, &cookie)

	C.fz_close_device(f.ctx, device)

	str := C.GoString(C.fz_string_from_buffer(f.ctx, buf))

	return str, nil
}

// Close closes the underlying fitz document.
func (f *Document) Close() error {
	C.fz_drop_document(f.ctx, f.doc)
	C.fz_drop_context(f.ctx)
	return nil
}
