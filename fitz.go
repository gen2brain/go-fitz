// Package fitz provides wrapper for the [MuPDF](http://mupdf.com/) fitz library
// that can extract pages from PDF and EPUB documents as images, text, html or svg.
package fitz

/*
#include <mupdf/fitz.h>
#include <stdlib.h>

const char *fz_version = FZ_VERSION;

fz_document *open_document(fz_context *ctx, const char *filename) {
	fz_document *doc;

	fz_try(ctx) {
		doc = fz_open_document(ctx, filename);
	}
	fz_catch(ctx) {
		return NULL;
	}

	return doc;
}

fz_document *open_document_with_stream(fz_context *ctx, const char *magic, fz_stream *stream) {
	fz_document *doc;

	fz_try(ctx) {
		doc = fz_open_document_with_stream(ctx, magic, stream);
	}
	fz_catch(ctx) {
		return NULL;
	}

	return doc;
}
*/
import "C"

import (
	"errors"
	"image"
	"io"
	"os"
	"path/filepath"
	"sync"
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
	ErrLoadOutline   = errors.New("fitz: cannot load outline")
)

// Document represents fitz document.
type Document struct {
	ctx    *C.struct_fz_context
	data   []byte // binds data to the Document lifecycle avoiding premature GC
	doc    *C.struct_fz_document
	mtx    sync.Mutex
	stream *C.fz_stream
}

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

	f.ctx = (*C.struct_fz_context)(unsafe.Pointer(C.fz_new_context_imp(nil, nil, C.FZ_STORE_UNLIMITED, C.fz_version)))
	if f.ctx == nil {
		err = ErrCreateContext
		return
	}

	C.fz_register_document_handlers(f.ctx)

	cfilename := C.CString(filename)
	defer C.free(unsafe.Pointer(cfilename))

	f.doc = C.open_document(f.ctx, cfilename)
	if f.doc == nil {
		err = ErrOpenDocument
		return
	}

	ret := C.fz_needs_password(f.ctx, f.doc)
	v := int(ret) != 0
	if v {
		err = ErrNeedsPassword
	}

	return
}

// NewFromMemory returns new fitz document from byte slice.
func NewFromMemory(b []byte) (f *Document, err error) {
	f = &Document{}

	f.ctx = (*C.struct_fz_context)(unsafe.Pointer(C.fz_new_context_imp(nil, nil, C.FZ_STORE_UNLIMITED, C.fz_version)))
	if f.ctx == nil {
		err = ErrCreateContext
		return
	}

	C.fz_register_document_handlers(f.ctx)

	stream := C.fz_open_memory(f.ctx, (*C.uchar)(&b[0]), C.size_t(len(b)))
	f.stream = C.fz_keep_stream(f.ctx, stream)

	if f.stream == nil {
		err = ErrOpenMemory
		return
	}

	magic := contentType(b)
	if magic == "" {
		err = ErrOpenMemory
		return
	}

	f.data = b

	cmagic := C.CString(magic)
	defer C.free(unsafe.Pointer(cmagic))

	f.doc = C.open_document_with_stream(f.ctx, cmagic, f.stream)
	if f.doc == nil {
		err = ErrOpenDocument
	}

	ret := C.fz_needs_password(f.ctx, f.doc)
	v := int(ret) != 0
	if v {
		err = ErrNeedsPassword
	}

	return
}

// NewFromReader returns new fitz document from io.Reader.
func NewFromReader(r io.Reader) (f *Document, err error) {
	b, e := io.ReadAll(r)
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
	return f.ImageDPI(pageNumber, 300.0)
}

// ImageDPI returns image for given page number and DPI.
func (f *Document) ImageDPI(pageNumber int, dpi float64) (image.Image, error) {
	f.mtx.Lock()
	defer f.mtx.Unlock()

	img := image.RGBA{}

	if pageNumber >= f.NumPage() {
		return nil, ErrPageMissing
	}

	page := C.fz_load_page(f.ctx, f.doc, C.int(pageNumber))
	defer C.fz_drop_page(f.ctx, page)

	var bounds C.fz_rect
	bounds = C.fz_bound_page(f.ctx, page)

	var ctm C.fz_matrix
	ctm = C.fz_scale(C.float(dpi/72), C.float(dpi/72))

	var bbox C.fz_irect
	bounds = C.fz_transform_rect(bounds, ctm)
	bbox = C.fz_round_rect(bounds)

	pixmap := C.fz_new_pixmap_with_bbox(f.ctx, C.fz_device_rgb(f.ctx), bbox, nil, 1)
	if pixmap == nil {
		return nil, ErrCreatePixmap
	}

	C.fz_clear_pixmap_with_value(f.ctx, pixmap, C.int(0xff))
	//defer C.fz_drop_pixmap(f.ctx, pixmap)

	device := C.fz_new_draw_device(f.ctx, ctm, pixmap)
	C.fz_enable_device_hints(f.ctx, device, C.FZ_NO_CACHE)
	defer C.fz_drop_device(f.ctx, device)

	drawMatrix := C.fz_identity
	C.fz_run_page(f.ctx, page, device, drawMatrix, nil)

	C.fz_close_device(f.ctx, device)

	pixels := C.fz_pixmap_samples(f.ctx, pixmap)
	if pixels == nil {
		return nil, ErrPixmapSamples
	}
	defer C.free(unsafe.Pointer(pixels))

	img.Pix = C.GoBytes(unsafe.Pointer(pixels), C.int(4*bbox.x1*bbox.y1))
	img.Rect = image.Rect(int(bbox.x0), int(bbox.y0), int(bbox.x1), int(bbox.y1))
	img.Stride = 4 * img.Rect.Max.X

	return &img, nil
}

// ImagePNG returns image for given page number as PNG bytes.
func (f *Document) ImagePNG(pageNumber int, dpi float64) ([]byte, error) {
	f.mtx.Lock()
	defer f.mtx.Unlock()

	if pageNumber >= f.NumPage() {
		return nil, ErrPageMissing
	}

	page := C.fz_load_page(f.ctx, f.doc, C.int(pageNumber))
	defer C.fz_drop_page(f.ctx, page)

	var bounds C.fz_rect
	bounds = C.fz_bound_page(f.ctx, page)

	var ctm C.fz_matrix
	ctm = C.fz_scale(C.float(dpi/72), C.float(dpi/72))

	var bbox C.fz_irect
	bounds = C.fz_transform_rect(bounds, ctm)
	bbox = C.fz_round_rect(bounds)

	pixmap := C.fz_new_pixmap_with_bbox(f.ctx, C.fz_device_rgb(f.ctx), bbox, nil, 1)
	if pixmap == nil {
		return nil, ErrCreatePixmap
	}

	C.fz_clear_pixmap_with_value(f.ctx, pixmap, C.int(0xff))
	//defer C.fz_drop_pixmap(f.ctx, pixmap)

	device := C.fz_new_draw_device(f.ctx, ctm, pixmap)
	C.fz_enable_device_hints(f.ctx, device, C.FZ_NO_CACHE)
	defer C.fz_drop_device(f.ctx, device)

	drawMatrix := C.fz_identity
	C.fz_run_page(f.ctx, page, device, drawMatrix, nil)

	C.fz_close_device(f.ctx, device)

	buf := C.fz_new_buffer_from_pixmap_as_png(f.ctx, pixmap, C.fz_default_color_params)
	defer C.fz_drop_buffer(f.ctx, buf)

	size := C.fz_buffer_storage(f.ctx, buf, nil)
	str := C.GoStringN(C.fz_string_from_buffer(f.ctx, buf), C.int(size))

	return []byte(str), nil
}

// Links returns slice of links for given page number.
func (f *Document) Links(pageNumber int) ([]Link, error) {
	f.mtx.Lock()
	defer f.mtx.Unlock()

	if pageNumber >= f.NumPage() {
		return nil, ErrPageMissing
	}

	page := C.fz_load_page(f.ctx, f.doc, C.int(pageNumber))
	defer C.fz_drop_page(f.ctx, page)

	links := C.fz_load_links(f.ctx, page)
	defer C.fz_drop_link(f.ctx, links)

	linkCount := 0
	for currLink := links; currLink != nil; currLink = currLink.next {
		linkCount++
	}

	if linkCount == 0 {
		return nil, nil
	}

	gLinks := make([]Link, linkCount)

	currLink := links
	for i := 0; i < linkCount; i++ {
		gLinks[i] = Link{
			URI: C.GoString(currLink.uri),
		}
		currLink = currLink.next
	}

	return gLinks, nil
}

// Text returns text for given page number.
func (f *Document) Text(pageNumber int) (string, error) {
	f.mtx.Lock()
	defer f.mtx.Unlock()

	if pageNumber >= f.NumPage() {
		return "", ErrPageMissing
	}

	page := C.fz_load_page(f.ctx, f.doc, C.int(pageNumber))
	defer C.fz_drop_page(f.ctx, page)

	var bounds C.fz_rect
	bounds = C.fz_bound_page(f.ctx, page)

	var ctm C.fz_matrix
	ctm = C.fz_scale(C.float(72.0/72), C.float(72.0/72))

	text := C.fz_new_stext_page(f.ctx, bounds)
	defer C.fz_drop_stext_page(f.ctx, text)

	var opts C.fz_stext_options
	opts.flags = 0

	device := C.fz_new_stext_device(f.ctx, text, &opts)
	C.fz_enable_device_hints(f.ctx, device, C.FZ_NO_CACHE)
	defer C.fz_drop_device(f.ctx, device)

	var cookie C.fz_cookie
	C.fz_run_page(f.ctx, page, device, ctm, &cookie)

	C.fz_close_device(f.ctx, device)

	buf := C.fz_new_buffer_from_stext_page(f.ctx, text)
	defer C.fz_drop_buffer(f.ctx, buf)

	str := C.GoString(C.fz_string_from_buffer(f.ctx, buf))

	return str, nil
}

// HTML returns html for given page number.
func (f *Document) HTML(pageNumber int, header bool) (string, error) {
	f.mtx.Lock()
	defer f.mtx.Unlock()

	if pageNumber >= f.NumPage() {
		return "", ErrPageMissing
	}

	page := C.fz_load_page(f.ctx, f.doc, C.int(pageNumber))
	defer C.fz_drop_page(f.ctx, page)

	var bounds C.fz_rect
	bounds = C.fz_bound_page(f.ctx, page)

	var ctm C.fz_matrix
	ctm = C.fz_scale(C.float(72.0/72), C.float(72.0/72))

	text := C.fz_new_stext_page(f.ctx, bounds)
	defer C.fz_drop_stext_page(f.ctx, text)

	var opts C.fz_stext_options
	opts.flags = C.FZ_STEXT_PRESERVE_IMAGES

	device := C.fz_new_stext_device(f.ctx, text, &opts)
	C.fz_enable_device_hints(f.ctx, device, C.FZ_NO_CACHE)
	defer C.fz_drop_device(f.ctx, device)

	var cookie C.fz_cookie
	C.fz_run_page(f.ctx, page, device, ctm, &cookie)

	C.fz_close_device(f.ctx, device)

	buf := C.fz_new_buffer(f.ctx, 1024)
	defer C.fz_drop_buffer(f.ctx, buf)

	out := C.fz_new_output_with_buffer(f.ctx, buf)
	defer C.fz_drop_output(f.ctx, out)

	if header {
		C.fz_print_stext_header_as_html(f.ctx, out)
	}
	C.fz_print_stext_page_as_html(f.ctx, out, text, C.int(pageNumber))
	if header {
		C.fz_print_stext_trailer_as_html(f.ctx, out)
	}

	str := C.GoString(C.fz_string_from_buffer(f.ctx, buf))

	return str, nil
}

// SVG returns svg document for given page number.
func (f *Document) SVG(pageNumber int) (string, error) {
	f.mtx.Lock()
	defer f.mtx.Unlock()

	if pageNumber >= f.NumPage() {
		return "", ErrPageMissing
	}

	page := C.fz_load_page(f.ctx, f.doc, C.int(pageNumber))
	defer C.fz_drop_page(f.ctx, page)

	var bounds C.fz_rect
	bounds = C.fz_bound_page(f.ctx, page)

	var ctm C.fz_matrix
	ctm = C.fz_scale(C.float(72.0/72), C.float(72.0/72))
	bounds = C.fz_transform_rect(bounds, ctm)

	buf := C.fz_new_buffer(f.ctx, 1024)
	defer C.fz_drop_buffer(f.ctx, buf)

	out := C.fz_new_output_with_buffer(f.ctx, buf)
	defer C.fz_drop_output(f.ctx, out)

	device := C.fz_new_svg_device(f.ctx, out, bounds.x1-bounds.x0, bounds.y1-bounds.y0, C.FZ_SVG_TEXT_AS_PATH, 1)
	C.fz_enable_device_hints(f.ctx, device, C.FZ_NO_CACHE)
	defer C.fz_drop_device(f.ctx, device)

	var cookie C.fz_cookie
	C.fz_run_page(f.ctx, page, device, ctm, &cookie)

	C.fz_close_device(f.ctx, device)

	str := C.GoString(C.fz_string_from_buffer(f.ctx, buf))

	return str, nil
}

// ToC returns the table of contents (also known as outline).
func (f *Document) ToC() ([]Outline, error) {
	data := make([]Outline, 0)

	outline := C.fz_load_outline(f.ctx, f.doc)
	if outline == nil {
		return nil, ErrLoadOutline
	}
	defer C.fz_drop_outline(f.ctx, outline)

	var walk func(outline *C.fz_outline, level int)

	walk = func(outline *C.fz_outline, level int) {
		for outline != nil {
			res := Outline{}
			res.Level = level
			res.Title = C.GoString(outline.title)
			res.URI = C.GoString(outline.uri)
			res.Page = int(outline.page.page)
			res.Top = float64(outline.y)
			data = append(data, res)

			if outline.down != nil {
				walk(outline.down, level+1)
			}
			outline = outline.next
		}
	}

	walk(outline, 1)
	return data, nil
}

// Metadata returns the map with standard metadata.
func (f *Document) Metadata() map[string]string {
	data := make(map[string]string)

	lookup := func(key string) string {
		ckey := C.CString(key)
		defer C.free(unsafe.Pointer(ckey))

		buf := make([]byte, 256)
		C.fz_lookup_metadata(f.ctx, f.doc, ckey, (*C.char)(unsafe.Pointer(&buf[0])), C.int(len(buf)))

		return string(buf)
	}

	data["format"] = lookup("format")
	data["encryption"] = lookup("encryption")
	data["title"] = lookup("info:Title")
	data["author"] = lookup("info:Author")
	data["subject"] = lookup("info:Subject")
	data["keywords"] = lookup("info:Keywords")
	data["creator"] = lookup("info:Creator")
	data["producer"] = lookup("info:Producer")
	data["creationDate"] = lookup("info:CreationDate")
	data["modDate"] = lookup("info:modDate")

	return data
}

// Bound gives the Bounds of a given Page in the document.
func (f *Document) Bound(pageNumber int) (image.Rectangle, error) {
	f.mtx.Lock()
	defer f.mtx.Unlock()

	if pageNumber >= f.NumPage() {
		return image.Rectangle{}, ErrPageMissing
	}

	page := C.fz_load_page(f.ctx, f.doc, C.int(pageNumber))
	defer C.fz_drop_page(f.ctx, page)

	var bounds C.fz_rect
	bounds = C.fz_bound_page(f.ctx, page)
	return image.Rect(int(bounds.x0), int(bounds.y0), int(bounds.x1), int(bounds.y1)), nil
}

// Close closes the underlying fitz document.
func (f *Document) Close() error {
	if f.stream != nil {
		C.fz_drop_stream(f.ctx, f.stream)
	}

	C.fz_drop_document(f.ctx, f.doc)
	C.fz_drop_context(f.ctx)

	f.data = nil

	return nil
}
