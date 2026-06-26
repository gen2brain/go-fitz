//go:build (!cgo || nocgo) && windows

package fitz

import "github.com/ebitengine/purego"

// Windows x64 passes >8-byte structs by pointer and returns them via sret, so by-value
// params are pointers (fz_bound_page gets a leading sret, fz_color_params packs into a uint32).
var (
	fzBoundPage                func(sret *fzRect, ctx *fzContext, page *fzPage) uintptr
	fzNewDrawDevice            func(ctx *fzContext, transform *fzMatrix, dest *fzPixmap) *fzDevice
	fzRunPageContents          func(ctx *fzContext, page *fzPage, dev *fzDevice, transform *fzMatrix, cookie *fzCookie)
	fzNewBufferFromPixmapAsPNG func(ctx *fzContext, pix *fzPixmap, params uint32) *fzBuffer
	fzNewStextPage             func(ctx *fzContext, mediabox *fzRect) *fzStextPage
)

func registerStructFuncs(lib uintptr) {
	purego.RegisterLibFunc(&fzBoundPage, lib, "fz_bound_page")
	purego.RegisterLibFunc(&fzNewDrawDevice, lib, "fz_new_draw_device")
	purego.RegisterLibFunc(&fzRunPageContents, lib, "fz_run_page_contents")
	purego.RegisterLibFunc(&fzNewBufferFromPixmapAsPNG, lib, "fz_new_buffer_from_pixmap_as_png")
	purego.RegisterLibFunc(&fzNewStextPage, lib, "fz_new_stext_page")
}

func boundPage(ctx *fzContext, page *fzPage) fzRect {
	var ret fzRect
	fzBoundPage(&ret, ctx, page)

	return ret
}

func newDrawDevice(ctx *fzContext, transform fzMatrix, dest *fzPixmap) *fzDevice {
	return fzNewDrawDevice(ctx, &transform, dest)
}

func runPageContents(ctx *fzContext, page *fzPage, dev *fzDevice, transform fzMatrix) {
	var cookie fzCookie
	fzRunPageContents(ctx, page, dev, &transform, &cookie)
}

func newBufferFromPixmapAsPNG(ctx *fzContext, pix *fzPixmap, params fzColorParams) *fzBuffer {
	packed := uint32(params.Ri) | uint32(params.Bp)<<8 | uint32(params.Op)<<16 | uint32(params.Opm)<<24

	return fzNewBufferFromPixmapAsPNG(ctx, pix, packed)
}

func newStextPage(ctx *fzContext, mediabox fzRect) *fzStextPage {
	return fzNewStextPage(ctx, &mediabox)
}
