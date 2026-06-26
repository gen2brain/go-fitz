//go:build (!cgo || nocgo) && !windows

package fitz

import "github.com/ebitengine/purego"

// Functions passing/returning MuPDF structs by value; purego handles these natively on SysV/AAPCS.
var (
	fzBoundPage                func(ctx *fzContext, page *fzPage) fzRect
	fzNewDrawDevice            func(ctx *fzContext, transform fzMatrix, dest *fzPixmap) *fzDevice
	fzRunPageContents          func(ctx *fzContext, page *fzPage, dev *fzDevice, transform fzMatrix, cookie *fzCookie)
	fzNewBufferFromPixmapAsPNG func(ctx *fzContext, pix *fzPixmap, params fzColorParams) *fzBuffer
	fzNewStextPage             func(ctx *fzContext, mediabox fzRect) *fzStextPage
)

func registerStructFuncs(lib uintptr) {
	purego.RegisterLibFunc(&fzBoundPage, lib, "fz_bound_page")
	purego.RegisterLibFunc(&fzNewDrawDevice, lib, "fz_new_draw_device")
	purego.RegisterLibFunc(&fzRunPageContents, lib, "fz_run_page_contents")
	purego.RegisterLibFunc(&fzNewBufferFromPixmapAsPNG, lib, "fz_new_buffer_from_pixmap_as_png")
	purego.RegisterLibFunc(&fzNewStextPage, lib, "fz_new_stext_page")
}

func boundPage(ctx *fzContext, page *fzPage) fzRect {
	return fzBoundPage(ctx, page)
}

func newDrawDevice(ctx *fzContext, transform fzMatrix, dest *fzPixmap) *fzDevice {
	return fzNewDrawDevice(ctx, transform, dest)
}

func runPageContents(ctx *fzContext, page *fzPage, dev *fzDevice, transform fzMatrix) {
	var cookie fzCookie
	fzRunPageContents(ctx, page, dev, transform, &cookie)
}

func newBufferFromPixmapAsPNG(ctx *fzContext, pix *fzPixmap, params fzColorParams) *fzBuffer {
	return fzNewBufferFromPixmapAsPNG(ctx, pix, params)
}

func newStextPage(ctx *fzContext, mediabox fzRect) *fzStextPage {
	return fzNewStextPage(ctx, mediabox)
}
