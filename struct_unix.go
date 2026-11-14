//go:build (!cgo || nocgo) && !windows

package fitz

import "github.com/ebitengine/purego"

// Functions passing/returning MuPDF structs by value; purego handles these natively on SysV/AAPCS.
var (
	fzBoundPage                func(ctx *fzContext, page *fzPage) fzRect
	fzNewDrawDevice            func(ctx *fzContext, transform fzMatrix, dest *fzPixmap) *fzDevice
	fzNewDisplayList           func(ctx *fzContext, mediabox fzRect) *fzDisplayList
	fzNewListDevice            func(ctx *fzContext, list *fzDisplayList) *fzDevice
	fzRunDisplayList           func(ctx *fzContext, list *fzDisplayList, dev *fzDevice, transform fzMatrix, scissor fzRect, cookie *fzCookie)
	fzRunPage                  func(ctx *fzContext, page *fzPage, dev *fzDevice, transform fzMatrix, cookie *fzCookie)
	fzNewBufferFromPixmapAsPNG func(ctx *fzContext, pix *fzPixmap, params fzColorParams) *fzBuffer
	fzNewStextPage             func(ctx *fzContext, mediabox fzRect) *fzStextPage
)

func registerStructFuncs(lib uintptr) {
	purego.RegisterLibFunc(&fzBoundPage, lib, "fz_bound_page")
	purego.RegisterLibFunc(&fzNewDrawDevice, lib, "fz_new_draw_device")
	purego.RegisterLibFunc(&fzNewDisplayList, lib, "fz_new_display_list")
	purego.RegisterLibFunc(&fzNewListDevice, lib, "fz_new_list_device")
	purego.RegisterLibFunc(&fzRunDisplayList, lib, "fz_run_display_list")
	purego.RegisterLibFunc(&fzDropDisplayList, lib, "fz_drop_display_list")
	purego.RegisterLibFunc(&fzRunPage, lib, "fz_run_page")
	purego.RegisterLibFunc(&fzNewBufferFromPixmapAsPNG, lib, "fz_new_buffer_from_pixmap_as_png")
	purego.RegisterLibFunc(&fzNewStextPage, lib, "fz_new_stext_page")
}

func boundPage(ctx *fzContext, page *fzPage) fzRect {
	return fzBoundPage(ctx, page)
}

func newDrawDevice(ctx *fzContext, transform fzMatrix, dest *fzPixmap) *fzDevice {
	return fzNewDrawDevice(ctx, transform, dest)
}

func newDisplayList(ctx *fzContext, mediabox fzRect) *fzDisplayList {
	return fzNewDisplayList(ctx, mediabox)
}

func newListDevice(ctx *fzContext, list *fzDisplayList) *fzDevice {
	return fzNewListDevice(ctx, list)
}

func runDisplayList(ctx *fzContext, list *fzDisplayList, dev *fzDevice, transform fzMatrix, scissor fzRect) {
	var cookie fzCookie
	fzRunDisplayList(ctx, list, dev, transform, scissor, &cookie)
}

func dropDisplayList(ctx *fzContext, list *fzDisplayList) {
	fzDropDisplayList(ctx, list)
}

func runPage(ctx *fzContext, page *fzPage, dev *fzDevice, transform fzMatrix) {
	var cookie fzCookie
	fzRunPage(ctx, page, dev, transform, &cookie)
}

func newBufferFromPixmapAsPNG(ctx *fzContext, pix *fzPixmap, params fzColorParams) *fzBuffer {
	return fzNewBufferFromPixmapAsPNG(ctx, pix, params)
}

func newStextPage(ctx *fzContext, mediabox fzRect) *fzStextPage {
	return fzNewStextPage(ctx, mediabox)
}
