//go:build !cgo || nocgo

package fitz

import (
	"fmt"
	"runtime"
	"unsafe"
)

// RenderPageBands renders directly into one bounded pixmap at a time. The
// purego backend does not currently expose MuPDF display-list bindings, so it
// uses direct page execution even when UseDisplayList is requested.
func (f *Document) RenderPageBands(
	pageNumber int,
	options BandOptions,
	consume func(Band) error,
) error {
	if consume == nil {
		return fmt.Errorf("fitz: nil band consumer")
	}
	if options.DPI <= 0 {
		return fmt.Errorf("fitz: invalid DPI: %v", options.DPI)
	}
	if options.BandHeight <= 0 {
		return fmt.Errorf("fitz: invalid band height: %d", options.BandHeight)
	}

	mupdf.Lock()
	defer mupdf.Unlock()

	if pageNumber < 0 || pageNumber >= f.numPage() {
		return ErrPageMissing
	}

	page := fzLoadPage(f.ctx, f.doc, pageNumber)
	if page == nil {
		return ErrLoadPage
	}
	defer fzDropPage(f.ctx, page)

	bounds := boundPage(f.ctx, page)
	ctm := scale(float32(options.DPI/72), float32(options.DPI/72))
	bounds = transformRect(bounds, ctm)
	bbox := roundRect(bounds)
	width := int(bbox.X1 - bbox.X0)
	totalHeight := int(bbox.Y1 - bbox.Y0)
	if width <= 0 || totalHeight <= 0 {
		return fmt.Errorf("fitz: invalid rendered page bounds")
	}

	for offset := 0; offset < totalHeight; offset += options.BandHeight {
		bandHeight := options.BandHeight
		if remaining := totalHeight - offset; remaining < bandHeight {
			bandHeight = remaining
		}

		pixmap := fzNewPixmap(f.ctx, fzDeviceRgb(f.ctx), width, bandHeight, nil, boolToInt(options.Alpha))
		if pixmap == nil {
			return ErrCreatePixmap
		}
		pixmap.X = bbox.X0
		pixmap.Y = bbox.Y0 + int32(offset)
		pixmap.W = int32(width)
		pixmap.H = int32(bandHeight)
		fzClearPixmapWithValue(f.ctx, pixmap, 0xff)

		device := newDrawDevice(f.ctx, ctm, pixmap)
		if device == nil {
			fzDropPixmap(f.ctx, pixmap)
			return ErrRunPageContents
		}
		fzEnableDeviceHints(f.ctx, device, fzNoCache)
		runPage(f.ctx, page, device, fzIdentity)
		fzCloseDevice(f.ctx, device)
		fzDropDevice(f.ctx, device)

		samples := fzPixmapSamples(f.ctx, pixmap)
		stride := int(pixmap.Stride)
		components := int(pixmap.N)
		if samples == nil || stride < width*components || components <= 0 {
			fzDropPixmap(f.ctx, pixmap)
			return ErrPixmapSamples
		}

		data := make([]byte, stride*bandHeight)
		copy(data, unsafe.Slice(samples, stride*bandHeight))
		fzDropPixmap(f.ctx, pixmap)

		if err := consume(Band{
			Y:          offset,
			Width:      width,
			Height:     bandHeight,
			Stride:     stride,
			Components: components,
			Samples:    data,
		}); err != nil {
			return err
		}
	}

	runtime.KeepAlive(f)
	return nil
}

func boolToInt(value bool) int {
	if value {
		return 1
	}
	return 0
}
