//go:build !cgo || nocgo

package fitz

import (
	"image"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"unsafe"

	"github.com/ebitengine/purego"
	"github.com/jupiterrider/ffi"
)

// Document represents fitz document.
type Document struct {
	ctx    *fzContext
	data   []byte // binds data to the Document lifecycle avoiding premature GC
	doc    *fzDocument
	mtx    sync.Mutex
	stream *fzStream
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

	f.ctx = fzNewContextImp(nil, nil, uint64(MaxStore), FzVersion)
	if f.ctx == nil {
		err = ErrCreateContext
		return
	}

	fzRegisterDocumentHandlers(f.ctx)

	f.doc = fzOpenDocument(f.ctx, filename)
	if f.doc == nil {
		err = ErrOpenDocument
		return
	}

	ret := fzNeedsPassword(f.ctx, f.doc)
	v := int(ret) != 0
	if v {
		err = ErrNeedsPassword
	}

	return
}

// NewFromMemory returns new fitz document from byte slice.
func NewFromMemory(b []byte) (f *Document, err error) {
	if len(b) = 0 {
		return nil, ErrEmptyBytes
	}
	f = &Document{}

	f.ctx = fzNewContextImp(nil, nil, uint64(MaxStore), FzVersion)
	if f.ctx == nil {
		err = ErrCreateContext
		return
	}

	fzRegisterDocumentHandlers(f.ctx)

	f.stream = fzOpenMemory(f.ctx, unsafe.SliceData(b), uint64(len(b)))
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

	f.doc = fzOpenDocumentWithStream(f.ctx, magic, f.stream)
	if f.doc == nil {
		err = ErrOpenDocument
	}

	ret := fzNeedsPassword(f.ctx, f.doc)
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
	return fzCountPages(f.ctx, f.doc)
}

// Image returns image for given page number.
func (f *Document) Image(pageNumber int) (*image.RGBA, error) {
	return f.ImageDPI(pageNumber, 300.0)
}

// ImageDPI returns image for given page number and DPI.
func (f *Document) ImageDPI(pageNumber int, dpi float64) (*image.RGBA, error) {
	f.mtx.Lock()
	defer f.mtx.Unlock()

	if pageNumber >= f.NumPage() {
		return nil, ErrPageMissing
	}

	page := fzLoadPage(f.ctx, f.doc, pageNumber)
	if page == nil {
		return nil, ErrLoadPage
	}

	defer fzDropPage(f.ctx, page)

	var bounds fzRect
	bounds = boundPage(f.ctx, page)

	var ctm fzMatrix
	ctm = scale(float32(dpi/72), float32(dpi/72))

	var bbox fzIRect
	bounds = transformRect(bounds, ctm)
	bbox = roundRect(bounds)

	pixmap := fzNewPixmap(f.ctx, fzDeviceRgb(f.ctx), int(bbox.X1), int(bbox.Y1), nil, 1)
	if pixmap == nil {
		return nil, ErrCreatePixmap
	}

	fzClearPixmapWithValue(f.ctx, pixmap, 0xff)
	defer fzDropPixmap(f.ctx, pixmap)

	device := newDrawDevice(f.ctx, ctm, pixmap)
	fzEnableDeviceHints(f.ctx, device, fzNoCache)
	defer fzDropDevice(f.ctx, device)

	runPageContents(f.ctx, page, device, fzIdentity)

	fzCloseDevice(f.ctx, device)

	pixels := fzPixmapSamples(f.ctx, pixmap)
	if pixels == nil {
		return nil, ErrPixmapSamples
	}

	img := image.NewRGBA(image.Rect(int(bbox.X0), int(bbox.Y0), int(bbox.X1), int(bbox.Y1)))
	copy(img.Pix, unsafe.Slice(pixels, 4*bbox.X1*bbox.Y1))

	return img, nil
}

// ImagePNG returns image for given page number as PNG bytes.
func (f *Document) ImagePNG(pageNumber int, dpi float64) ([]byte, error) {
	f.mtx.Lock()
	defer f.mtx.Unlock()

	if pageNumber >= f.NumPage() {
		return nil, ErrPageMissing
	}

	page := fzLoadPage(f.ctx, f.doc, pageNumber)
	if page == nil {
		return nil, ErrLoadPage
	}

	defer fzDropPage(f.ctx, page)

	var bounds fzRect
	bounds = boundPage(f.ctx, page)

	var ctm fzMatrix
	ctm = scale(float32(dpi/72), float32(dpi/72))

	var bbox fzIRect
	bounds = transformRect(bounds, ctm)
	bbox = roundRect(bounds)

	pixmap := fzNewPixmap(f.ctx, fzDeviceRgb(f.ctx), int(bbox.X1), int(bbox.Y1), nil, 1)
	if pixmap == nil {
		return nil, ErrCreatePixmap
	}

	fzClearPixmapWithValue(f.ctx, pixmap, 0xff)
	defer fzDropPixmap(f.ctx, pixmap)

	device := newDrawDevice(f.ctx, ctm, pixmap)
	fzEnableDeviceHints(f.ctx, device, fzNoCache)
	defer fzDropDevice(f.ctx, device)

	runPageContents(f.ctx, page, device, fzIdentity)

	fzCloseDevice(f.ctx, device)

	params := fzColorParams{1, 1, 0, 0}
	buf := newBufferFromPixmapAsPNG(f.ctx, pixmap, params)
	defer fzDropBuffer(f.ctx, buf)

	size := fzBufferStorage(f.ctx, buf, nil)

	ret := make([]byte, size)
	copy(ret, unsafe.Slice(fzStringFromBuffer(f.ctx, buf), size))

	return ret, nil
}

// Links returns slice of links for given page number.
func (f *Document) Links(pageNumber int) ([]Link, error) {
	f.mtx.Lock()
	defer f.mtx.Unlock()

	if pageNumber >= f.NumPage() {
		return nil, ErrPageMissing
	}

	page := fzLoadPage(f.ctx, f.doc, pageNumber)
	if page == nil {
		return nil, ErrLoadPage
	}

	defer fzDropPage(f.ctx, page)

	links := fzLoadLinks(f.ctx, page)
	defer fzDropLink(f.ctx, links)

	linkCount := 0
	for currLink := links; currLink != nil; currLink = currLink.Next {
		linkCount++
	}

	if linkCount == 0 {
		return nil, nil
	}

	gLinks := make([]Link, linkCount)

	currLink := links
	for i := 0; i < linkCount; i++ {
		gLinks[i] = Link{
			URI: bytePtrToString((*uint8)(unsafe.Pointer(currLink.Uri))),
		}
		currLink = currLink.Next
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

	page := fzLoadPage(f.ctx, f.doc, pageNumber)
	if page == nil {
		return "", ErrLoadPage
	}

	defer fzDropPage(f.ctx, page)

	var bounds fzRect
	bounds = boundPage(f.ctx, page)

	var ctm fzMatrix
	ctm = scale(float32(72.0/72), float32(72.0/72))

	text := newStextPage(f.ctx, bounds)
	defer fzDropStextPage(f.ctx, text)

	var opts fzStextOptions
	opts.Flags = 0

	device := fzNewStextDevice(f.ctx, text, &opts)
	fzEnableDeviceHints(f.ctx, device, fzNoCache)
	defer fzDropDevice(f.ctx, device)

	runPageContents(f.ctx, page, device, ctm)

	fzCloseDevice(f.ctx, device)

	buf := fzNewBufferFromStextPage(f.ctx, text)
	defer fzDropBuffer(f.ctx, buf)

	ret := fzStringFromBuffer(f.ctx, buf)

	return bytePtrToString(ret), nil
}

// HTML returns html for given page number.
func (f *Document) HTML(pageNumber int, header bool) (string, error) {
	f.mtx.Lock()
	defer f.mtx.Unlock()

	if pageNumber >= f.NumPage() {
		return "", ErrPageMissing
	}

	page := fzLoadPage(f.ctx, f.doc, pageNumber)
	if page == nil {
		return "", ErrLoadPage
	}

	defer fzDropPage(f.ctx, page)

	var bounds fzRect
	bounds = boundPage(f.ctx, page)

	var ctm fzMatrix
	ctm = scale(float32(72.0/72), float32(72.0/72))

	text := newStextPage(f.ctx, bounds)
	defer fzDropStextPage(f.ctx, text)

	var opts fzStextOptions
	opts.Flags = fzStextPreserveImages

	device := fzNewStextDevice(f.ctx, text, &opts)
	fzEnableDeviceHints(f.ctx, device, fzNoCache)
	defer fzDropDevice(f.ctx, device)

	runPageContents(f.ctx, page, device, ctm)

	fzCloseDevice(f.ctx, device)

	buf := fzNewBuffer(f.ctx, 1024)
	defer fzDropBuffer(f.ctx, buf)

	out := fzNewOutputWithBuffer(f.ctx, buf)
	defer fzDropOutput(f.ctx, out)

	if header {
		fzPrintStextHeaderAsHTML(f.ctx, out)
	}
	fzPrintStextPageAsHTML(f.ctx, out, text, pageNumber)
	if header {
		fzPrintStextTrailerAsHTML(f.ctx, out)
	}

	fzCloseOutput(f.ctx, out)

	ret := fzStringFromBuffer(f.ctx, buf)

	return bytePtrToString(ret), nil
}

// SVG returns svg document for given page number.
func (f *Document) SVG(pageNumber int) (string, error) {
	f.mtx.Lock()
	defer f.mtx.Unlock()

	if pageNumber >= f.NumPage() {
		return "", ErrPageMissing
	}

	page := fzLoadPage(f.ctx, f.doc, pageNumber)
	if page == nil {
		return "", ErrLoadPage
	}

	defer fzDropPage(f.ctx, page)

	var bounds fzRect
	bounds = boundPage(f.ctx, page)

	var ctm fzMatrix
	ctm = scale(float32(72.0/72), float32(72.0/72))
	bounds = transformRect(bounds, ctm)

	buf := fzNewBuffer(f.ctx, 1024)
	defer fzDropBuffer(f.ctx, buf)

	out := fzNewOutputWithBuffer(f.ctx, buf)
	defer fzDropOutput(f.ctx, out)

	device := newSvgDevice(f.ctx, out, bounds.X1-bounds.X0, bounds.Y1-bounds.Y0, fzSvgTextAsPath, 1)
	fzEnableDeviceHints(f.ctx, device, fzNoCache)
	defer fzDropDevice(f.ctx, device)

	runPageContents(f.ctx, page, device, ctm)

	fzCloseDevice(f.ctx, device)
	fzCloseOutput(f.ctx, out)

	ret := fzStringFromBuffer(f.ctx, buf)

	return bytePtrToString(ret), nil
}

// ToC returns the table of contents (also known as outline).
func (f *Document) ToC() ([]Outline, error) {
	data := make([]Outline, 0)

	outline := fzLoadOutline(f.ctx, f.doc)
	if outline == nil {
		return nil, ErrLoadOutline
	}

	defer fzDropOutline(f.ctx, outline)

	var walk func(outline *fzOutline, level int)

	walk = func(outline *fzOutline, level int) {
		for outline != nil {
			res := Outline{}
			res.Level = level
			res.Title = bytePtrToString((*uint8)(unsafe.Pointer(outline.Title)))
			res.URI = bytePtrToString((*uint8)(unsafe.Pointer(outline.Uri)))
			res.Page = int(outline.Page.Page)
			res.Top = float64(outline.Y)
			data = append(data, res)

			if outline.Down != nil {
				walk(outline.Down, level+1)
			}
			outline = outline.Next
		}
	}

	walk(outline, 1)

	return data, nil
}

// Metadata returns the map with standard metadata.
func (f *Document) Metadata() map[string]string {
	data := make(map[string]string)

	lookup := func(key string) string {
		buf := make([]byte, 256)
		fzLookupMetadata(f.ctx, f.doc, key, unsafe.SliceData(buf), len(buf))

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

	page := fzLoadPage(f.ctx, f.doc, pageNumber)
	if page == nil {
		return image.Rectangle{}, ErrLoadPage
	}

	defer fzDropPage(f.ctx, page)

	var bounds fzRect
	bounds = boundPage(f.ctx, page)

	return image.Rect(int(bounds.X0), int(bounds.Y0), int(bounds.X1), int(bounds.Y1)), nil
}

// Close closes the underlying fitz document.
func (f *Document) Close() error {
	if f.stream != nil {
		fzDropStream(f.ctx, f.stream)
	}

	fzDropDocument(f.ctx, f.doc)
	fzDropContext(f.ctx)

	f.data = nil

	return nil
}

var (
	libmupdf uintptr

	fzBoundPage                *bundle
	fzTransformRect            *bundle
	fzRoundRect                *bundle
	fzScale                    *bundle
	fzNewDrawDevice            *bundle
	fzRunPageContents          *bundle
	fzNewBufferFromPixmapAsPNG *bundle
	fzNewStextPage             *bundle
	fzNewSvgDevice             *bundle

	fzNewContextImp            func(alloc *fzAllocContext, locks *fzLocksContext, maxStore uint64, version string) *fzContext
	fzDropContext              func(ctx *fzContext)
	fzOpenDocument             func(ctx *fzContext, filename string) *fzDocument
	fzOpenDocumentWithStream   func(ctx *fzContext, magic string, stream *fzStream) *fzDocument
	fzOpenMemory               func(ctx *fzContext, data *uint8, len uint64) *fzStream
	fzDropStream               func(ctx *fzContext, stm *fzStream)
	fzRegisterDocumentHandlers func(ctx *fzContext)
	fzNeedsPassword            func(ctx *fzContext, doc *fzDocument) int
	fzDropDocument             func(ctx *fzContext, doc *fzDocument)
	fzCountPages               func(ctx *fzContext, doc *fzDocument) int
	fzLoadPage                 func(ctx *fzContext, doc *fzDocument, number int) *fzPage
	fzDropPage                 func(ctx *fzContext, page *fzPage)
	fzNewPixmap                func(ctx *fzContext, colorspace *fzColorspace, w, h int, seps *fzSeparations, alpha int) *fzPixmap
	fzDropPixmap               func(ctx *fzContext, pix *fzPixmap)
	fzPixmapSamples            func(ctx *fzContext, pix *fzPixmap) *uint8
	fzClearPixmapWithValue     func(ctx *fzContext, pix *fzPixmap, value int)
	fzEnableDeviceHints        func(ctx *fzContext, dev *fzDevice, hints int)
	fzDropDevice               func(ctx *fzContext, dev *fzDevice)
	fzCloseDevice              func(ctx *fzContext, dev *fzDevice)
	fzDeviceRgb                func(ctx *fzContext) *fzColorspace
	fzNewBuffer                func(ctx *fzContext, size uint64) *fzBuffer
	fzDropBuffer               func(ctx *fzContext, buf *fzBuffer)
	fzBufferStorage            func(ctx *fzContext, buf *fzBuffer, data **uint8) uint64
	fzStringFromBuffer         func(ctx *fzContext, buf *fzBuffer) *uint8
	fzLoadLinks                func(ctx *fzContext, page *fzPage) *fzLink
	fzDropLink                 func(ctx *fzContext, link *fzLink)
	fzDropStextPage            func(ctx *fzContext, page *fzStextPage)
	fzNewStextDevice           func(ctx *fzContext, page *fzStextPage, options *fzStextOptions) *fzDevice
	fzNewBufferFromStextPage   func(ctx *fzContext, page *fzStextPage) *fzBuffer
	fzLookupMetadata           func(ctx *fzContext, doc *fzDocument, key string, buf *uint8, size int) int
	fzLoadOutline              func(ctx *fzContext, doc *fzDocument) *fzOutline
	fzDropOutline              func(ctx *fzContext, outline *fzOutline)
	fzNewOutputWithBuffer      func(ctx *fzContext, buf *fzBuffer) *fzOutput
	fzDropOutput               func(ctx *fzContext, out *fzOutput)
	fzCloseOutput              func(ctx *fzContext, out *fzOutput)
	fzPrintStextPageAsHTML     func(ctx *fzContext, out *fzOutput, page *fzStextPage, id int)
	fzPrintStextHeaderAsHTML   func(ctx *fzContext, out *fzOutput)
	fzPrintStextTrailerAsHTML  func(ctx *fzContext, out *fzOutput)
)

func init() {
	libmupdf = loadLibrary()

	if os.Getenv("FZ_VERSION") != "" {
		FzVersion = os.Getenv("FZ_VERSION")
	}

	fzBoundPage = newBundle("fz_bound_page", &typeFzRect, &ffi.TypePointer, &ffi.TypePointer)
	fzTransformRect = newBundle("fz_transform_rect", &typeFzRect, &typeFzRect, &typeFzMatrix)
	fzRoundRect = newBundle("fz_round_rect", &typeFzIRect, &typeFzRect)
	fzScale = newBundle("fz_scale", &typeFzMatrix, &ffi.TypeFloat, &ffi.TypeFloat)
	fzNewDrawDevice = newBundle("fz_new_draw_device", &ffi.TypePointer, &ffi.TypePointer, &typeFzMatrix, &ffi.TypePointer)
	fzRunPageContents = newBundle("fz_run_page_contents", &ffi.TypeVoid, &ffi.TypePointer, &ffi.TypePointer, &ffi.TypePointer, &typeFzMatrix, &ffi.TypePointer)
	fzNewBufferFromPixmapAsPNG = newBundle("fz_new_buffer_from_pixmap_as_png", &ffi.TypePointer, &ffi.TypePointer, &ffi.TypePointer, &typeFzColorParams)
	fzNewStextPage = newBundle("fz_new_stext_page", &ffi.TypePointer, &ffi.TypePointer, &typeFzRect)
	fzNewSvgDevice = newBundle("fz_new_svg_device", &ffi.TypePointer, &ffi.TypePointer, &ffi.TypePointer, &ffi.TypeFloat, &ffi.TypeFloat, &ffi.TypeSint32, &ffi.TypeSint32)

	purego.RegisterLibFunc(&fzNewContextImp, libmupdf, "fz_new_context_imp")
	purego.RegisterLibFunc(&fzDropContext, libmupdf, "fz_drop_context")
	purego.RegisterLibFunc(&fzOpenDocument, libmupdf, "fz_open_document")
	purego.RegisterLibFunc(&fzOpenDocumentWithStream, libmupdf, "fz_open_document_with_stream")
	purego.RegisterLibFunc(&fzOpenMemory, libmupdf, "fz_open_memory")
	purego.RegisterLibFunc(&fzDropStream, libmupdf, "fz_drop_stream")
	purego.RegisterLibFunc(&fzRegisterDocumentHandlers, libmupdf, "fz_register_document_handlers")
	purego.RegisterLibFunc(&fzNeedsPassword, libmupdf, "fz_needs_password")
	purego.RegisterLibFunc(&fzDropDocument, libmupdf, "fz_drop_document")
	purego.RegisterLibFunc(&fzCountPages, libmupdf, "fz_count_pages")
	purego.RegisterLibFunc(&fzLoadPage, libmupdf, "fz_load_page")
	purego.RegisterLibFunc(&fzDropPage, libmupdf, "fz_drop_page")
	purego.RegisterLibFunc(&fzNewPixmap, libmupdf, "fz_new_pixmap")
	purego.RegisterLibFunc(&fzDropPixmap, libmupdf, "fz_drop_pixmap")
	purego.RegisterLibFunc(&fzPixmapSamples, libmupdf, "fz_pixmap_samples")
	purego.RegisterLibFunc(&fzClearPixmapWithValue, libmupdf, "fz_clear_pixmap_with_value")
	purego.RegisterLibFunc(&fzEnableDeviceHints, libmupdf, "fz_enable_device_hints")
	purego.RegisterLibFunc(&fzDropDevice, libmupdf, "fz_drop_device")
	purego.RegisterLibFunc(&fzCloseDevice, libmupdf, "fz_close_device")
	purego.RegisterLibFunc(&fzDeviceRgb, libmupdf, "fz_device_rgb")
	purego.RegisterLibFunc(&fzNewBuffer, libmupdf, "fz_new_buffer")
	purego.RegisterLibFunc(&fzDropBuffer, libmupdf, "fz_drop_buffer")
	purego.RegisterLibFunc(&fzBufferStorage, libmupdf, "fz_buffer_storage")
	purego.RegisterLibFunc(&fzStringFromBuffer, libmupdf, "fz_string_from_buffer")
	purego.RegisterLibFunc(&fzLoadLinks, libmupdf, "fz_load_links")
	purego.RegisterLibFunc(&fzDropLink, libmupdf, "fz_drop_link")
	purego.RegisterLibFunc(&fzDropStextPage, libmupdf, "fz_drop_stext_page")
	purego.RegisterLibFunc(&fzNewStextDevice, libmupdf, "fz_new_stext_device")
	purego.RegisterLibFunc(&fzNewBufferFromStextPage, libmupdf, "fz_new_buffer_from_stext_page")
	purego.RegisterLibFunc(&fzLookupMetadata, libmupdf, "fz_lookup_metadata")
	purego.RegisterLibFunc(&fzLoadOutline, libmupdf, "fz_load_outline")
	purego.RegisterLibFunc(&fzDropOutline, libmupdf, "fz_drop_outline")
	purego.RegisterLibFunc(&fzNewOutputWithBuffer, libmupdf, "fz_new_output_with_buffer")
	purego.RegisterLibFunc(&fzDropOutput, libmupdf, "fz_drop_output")
	purego.RegisterLibFunc(&fzCloseOutput, libmupdf, "fz_close_output")
	purego.RegisterLibFunc(&fzPrintStextPageAsHTML, libmupdf, "fz_print_stext_page_as_html")
	purego.RegisterLibFunc(&fzPrintStextHeaderAsHTML, libmupdf, "fz_print_stext_header_as_html")
	purego.RegisterLibFunc(&fzPrintStextTrailerAsHTML, libmupdf, "fz_print_stext_trailer_as_html")

	ver := version()
	if ver != "" {
		FzVersion = ver
	}
}

func version() string {
	if fzNewContextImp(nil, nil, uint64(MaxStore), FzVersion) != nil {
		return FzVersion
	}

	s := strings.Split(FzVersion, ".")
	v := strings.Join(s[:len(s)-1], ".")

	for x := 10; x >= 0; x-- {
		ver := v + "." + strconv.Itoa(x)
		if ver == FzVersion {
			continue
		}

		if fzNewContextImp(nil, nil, uint64(MaxStore), ver) != nil {
			return ver
		}
	}

	return ""
}

type bundle struct {
	sym uintptr
	cif ffi.Cif
}

func (b *bundle) call(rValue unsafe.Pointer, aValues ...unsafe.Pointer) {
	ffi.Call(&b.cif, b.sym, rValue, aValues...)
}

func newBundle(name string, rType *ffi.Type, aTypes ...*ffi.Type) *bundle {
	b := new(bundle)
	var err error

	if b.sym, err = purego.Dlsym(libmupdf, name); err != nil {
		panic(err)
	}

	nArgs := uint32(len(aTypes))

	if status := ffi.PrepCif(&b.cif, ffi.DefaultAbi, nArgs, rType, aTypes...); status != ffi.OK {
		panic(status)
	}

	return b
}

var typeFzRect = ffi.Type{Type: ffi.Struct, Elements: &[]*ffi.Type{&ffi.TypeFloat, &ffi.TypeFloat, &ffi.TypeFloat, &ffi.TypeFloat, nil}[0]}
var typeFzIRect = ffi.Type{Type: ffi.Struct, Elements: &[]*ffi.Type{&ffi.TypeSint32, &ffi.TypeSint32, &ffi.TypeSint32, &ffi.TypeSint32, nil}[0]}
var typeFzMatrix = ffi.Type{Type: ffi.Struct, Elements: &[]*ffi.Type{&ffi.TypeFloat, &ffi.TypeFloat, &ffi.TypeFloat, &ffi.TypeFloat, &ffi.TypeFloat, &ffi.TypeFloat, nil}[0]}
var typeFzColorParams = ffi.Type{Type: ffi.Struct, Elements: &[]*ffi.Type{&ffi.TypeUint8, &ffi.TypeUint8, &ffi.TypeUint8, &ffi.TypeUint8, nil}[0]}

func boundPage(ctx *fzContext, page *fzPage) fzRect {
	var ret fzRect
	fzBoundPage.call(unsafe.Pointer(&ret), unsafe.Pointer(&ctx), unsafe.Pointer(&page))

	return ret
}

func transformRect(rect fzRect, m fzMatrix) fzRect {
	var ret fzRect
	fzTransformRect.call(unsafe.Pointer(&ret), unsafe.Pointer(&rect), unsafe.Pointer(&m))

	return ret
}

func roundRect(rect fzRect) fzIRect {
	var ret fzIRect
	fzRoundRect.call(unsafe.Pointer(&ret), unsafe.Pointer(&rect))

	return ret
}

func scale(sx, sy float32) fzMatrix {
	var ret fzMatrix
	fzScale.call(unsafe.Pointer(&ret), unsafe.Pointer(&sx), unsafe.Pointer(&sy))

	return ret
}

func newDrawDevice(ctx *fzContext, transform fzMatrix, dest *fzPixmap) *fzDevice {
	var ret *fzDevice
	fzNewDrawDevice.call(unsafe.Pointer(&ret), unsafe.Pointer(&ctx), unsafe.Pointer(&transform), unsafe.Pointer(&dest))

	return ret
}

func runPageContents(ctx *fzContext, page *fzPage, dev *fzDevice, transform fzMatrix) {
	var cookie fzCookie
	fzRunPageContents.call(nil, unsafe.Pointer(&ctx), unsafe.Pointer(&page), unsafe.Pointer(&dev), unsafe.Pointer(&transform), unsafe.Pointer(&cookie))
}

func newBufferFromPixmapAsPNG(ctx *fzContext, pix *fzPixmap, params fzColorParams) *fzBuffer {
	var ret *fzBuffer
	fzNewBufferFromPixmapAsPNG.call(unsafe.Pointer(&ret), unsafe.Pointer(&ctx), unsafe.Pointer(&pix), unsafe.Pointer(&params))

	return ret
}
func newStextPage(ctx *fzContext, mediabox fzRect) *fzStextPage {
	var ret *fzStextPage
	fzNewStextPage.call(unsafe.Pointer(&ret), unsafe.Pointer(&ctx), unsafe.Pointer(&mediabox))

	return ret
}

func newSvgDevice(ctx *fzContext, out *fzOutput, pageWidth, pageHeight float32, textFormat, reuseImages int) *fzDevice {
	var ret *fzDevice
	fzNewSvgDevice.call(unsafe.Pointer(&ret), unsafe.Pointer(&ctx), unsafe.Pointer(&out), unsafe.Pointer(&pageWidth), unsafe.Pointer(&pageHeight), unsafe.Pointer(&textFormat), unsafe.Pointer(&reuseImages))

	return ret
}

const (
	fzNoCache             = 2
	fzStextPreserveImages = 4
	fzSvgTextAsPath       = 0
)

var fzIdentity = fzMatrix{A: 1, B: 0, C: 0, D: 1, E: 0, F: 0}

type fzContext struct {
	User          *byte
	Alloc         fzAllocContext
	Locks         fzLocksContext
	Error         fzErrorContext
	Warn          fzWarnContext
	Aa            fzAaContext
	Seed48        [7]uint16
	IccEnabled    int32
	ThrowOnRepair int32
	Handler       *fzDocumentHandlerContext
	Archive       *fzArchiveHandlerContext
	Style         *fzStyleContext
	Tuning        *fzTuningContext
	StdDbg        *fzOutput
	Font          *fzFontContext
	Colorspace    *fzColorspaceContext
	Store         *fzStore
	GlyphCache    *fzGlyphCache
}

type fzDocument struct {
	Refs                 int32
	DropDocument         *[0]byte
	NeedsPassword        *[0]byte
	AuthenticatePassword *[0]byte
	HasPermission        *[0]byte
	LoadOutline          *[0]byte
	OutlineIterator      *[0]byte
	Layout               *[0]byte
	MakeBookmark         *[0]byte
	LookupBookmark       *[0]byte
	ResolveLinkDest      *[0]byte
	FormatLinkUri        *[0]byte
	CountChapters        *[0]byte
	CountPages           *[0]byte
	LoadPage             *[0]byte
	PageLabel            *[0]byte
	LookupMetadata       *[0]byte
	SetMetadata          *[0]byte
	GetOutputIntent      *[0]byte
	OutputAccelerator    *[0]byte
	RunStructure         *[0]byte
	AsPdf                *[0]byte
	DidLayout            int32
	IsReflowable         int32
	Open                 *fzPage
}

type fzOutline struct {
	Refs  int32
	Title *int8
	Uri   *int8
	Page  fzLocation
	X     float32
	Y     float32
	Next  *fzOutline
	Down  *fzOutline
	Open  int32
	_     [4]byte
}

type fzPage struct {
	Refs               int32
	Doc                *fzDocument
	Chapter            int32
	Number             int32
	Incomplete         int32
	DropPage           *[0]byte
	BoundPage          *[0]byte
	RunPageContents    *[0]byte
	RunPageAnnots      *[0]byte
	RunPageWidgets     *[0]byte
	LoadLinks          *[0]byte
	PagePresentation   *[0]byte
	ControlSeparation  *[0]byte
	SeparationDisabled *[0]byte
	Separations        *[0]byte
	Overprint          *[0]byte
	CreateLink         *[0]byte
	DeleteLink         *[0]byte
	Prev               **fzPage
	Next               *fzPage
}

type fzOutput struct {
	State    *byte
	Write    *[0]byte
	Seek     *[0]byte
	Tell     *[0]byte
	Close    *[0]byte
	Drop     *[0]byte
	Reset    *[0]byte
	Stream   *[0]byte
	Truncate *[0]byte
	Closed   int32
	Bp       *int8
	Wp       *int8
	Ep       *int8
	Buffered int32
	Bits     int32
}

type fzLocation struct {
	Chapter int32
	Page    int32
}

type fzStream struct {
	Refs        int32
	Error       int32
	Eof         int32
	Progressive int32
	Pos         int64
	Avail       int32
	Bits        int32
	Rp          *uint8
	Wp          *uint8
	State       *byte
	Next        *[0]byte
	Drop        *[0]byte
	Seek        *[0]byte
}

type fzRect struct {
	X0 float32
	Y0 float32
	X1 float32
	Y1 float32
}

type fzIRect struct {
	X0 int32
	Y0 int32
	X1 int32
	Y1 int32
}

type fzMatrix struct {
	A float32
	B float32
	C float32
	D float32
	E float32
	F float32
}

type fzCookie struct {
	Abort      int32
	Progress   int32
	Max        uint64
	Errors     int32
	Incomplete int32
}

type fzDevice struct {
	Refs                  int32
	Hints                 int32
	Flags                 int32
	CloseDevice           *[0]byte
	DropDevice            *[0]byte
	FillPath              *[0]byte
	StrokePath            *[0]byte
	ClipPath              *[0]byte
	ClipStrokePath        *[0]byte
	FillText              *[0]byte
	StrokeText            *[0]byte
	ClipText              *[0]byte
	ClipStrokeText        *[0]byte
	IgnoreText            *[0]byte
	FillShade             *[0]byte
	FillImage             *[0]byte
	FillImageMask         *[0]byte
	ClipImageMask         *[0]byte
	PopClip               *[0]byte
	BeginMask             *[0]byte
	EndMask               *[0]byte
	BeginGroup            *[0]byte
	EndGroup              *[0]byte
	BeginTile             *[0]byte
	EndTile               *[0]byte
	RenderFlags           *[0]byte
	SetDefaultColorspaces *[0]byte
	BeginLayer            *[0]byte
	EndLayer              *[0]byte
	BeginStructure        *[0]byte
	EndStructure          *[0]byte
	BeginMetatext         *[0]byte
	EndMetatext           *[0]byte
	D1Rect                fzRect
	ContainerLen          int32
	ContainerCap          int32
	Container             *fzDeviceContainerStack
}

type fzColorspace struct {
	Storable fzKeyStorable
	Type     uint32
	Flags    int32
	N        int32
	Name     *int8
	U        [288]byte
}

type fzStorable struct {
	Refs      int32
	Drop      *[0]byte
	Droppable *[0]byte
}

type fzKeyStorable struct {
	Storable fzStorable
	KeyRefs  int16
	_        [6]byte
}

type fzPixmap struct {
	Storable   fzStorable
	X          int32
	Y          int32
	W          int32
	H          int32
	N          uint8
	S          uint8
	Alpha      uint8
	Flags      uint8
	Stride     int64
	Seps       *fzSeparations
	Xres       int32
	Yres       int32
	Colorspace *fzColorspace
	Samples    *uint8
	Underlying *fzPixmap
}

type fzColorParams struct {
	Ri  uint8
	Bp  uint8
	Op  uint8
	Opm uint8
}

type fzBuffer struct {
	Refs   int32
	Data   *uint8
	Cap    uint64
	Len    uint64
	Bits   int32
	Shared int32
}

type fzLink struct {
	Refs   int32
	Next   *fzLink
	Rect   fzRect
	Uri    *int8
	RectFn *[0]byte
	UriFn  *[0]byte
	Drop   *[0]byte
}

type fzStextPage struct {
	Pool       *fzPool
	Mediabox   fzRect
	FirstBlock *fzStextBlock
	LastBlock  *fzStextBlock
}

type fzStextOptions struct {
	Flags int32
	Scale float32
}

type fzStextBlock struct {
	Type int32
	Bbox fzRect
	_    [4]byte
	U    [32]byte
	Prev *fzStextBlock
	Next *fzStextBlock
}

type fzDeviceContainerStack struct {
	Scissor fzRect
	Type    int32
	User    int32
}

type fzAllocContext struct {
	User    *byte
	Malloc  *[0]byte
	Realloc *[0]byte
	Free    *[0]byte
}

type fzLocksContext struct {
	User   *byte
	Lock   *[0]byte
	Unlock *[0]byte
}

type fzErrorContext struct {
	Top       *fzErrorStackSlot
	Stack     [256]fzErrorStackSlot
	Padding   fzErrorStackSlot
	StackBase *fzErrorStackSlot
	ErrCode   int32
	ErrNum    int32
	PrintUser *byte
	Print     *[0]byte
	Message   [256]int8
}

type fzWarnContext struct {
	User    *byte
	Print   *[0]byte
	Count   int32
	Message [256]int8
	_       [4]byte
}

type fzAaContext struct {
	Hscale       int32
	Vscale       int32
	Scale        int32
	Bits         int32
	TextBits     int32
	MinLineWidth float32
}

type fzErrorStackSlot struct {
	Buffer  [1]int32
	State   int32
	Code    int32
	Padding [24]int8
}

type fzFontContext struct{}
type fzColorspaceContext struct{}
type fzTuningContext struct{}
type fzStyleContext struct{}
type fzDocumentHandlerContext struct{}
type fzArchiveHandlerContext struct{}
type fzStore struct{}
type fzGlyphCache struct{}
type fzSeparations struct{}
type fzPool struct{}
