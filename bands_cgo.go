//go:build cgo && !nocgo

package fitz

/*
#include <mupdf/fitz.h>
#include <stdlib.h>

static int band_page_bounds(fz_context *ctx, fz_page *page, float dpi, fz_irect *out) {
	fz_rect bounds = fz_bound_page(ctx, page);
	fz_matrix ctm = fz_scale(dpi / 72.0f, dpi / 72.0f);
	*out = fz_round_rect(fz_transform_rect(bounds, ctm));
	return out->x1 > out->x0 && out->y1 > out->y0;
}

static fz_page *load_page_safe(fz_context *ctx, fz_document *doc, int number) {
	fz_page *page = NULL;
	fz_try(ctx) {
		page = fz_load_page(ctx, doc, number);
	}
	fz_catch(ctx) {
		return NULL;
	}
	return page;
}

static fz_pixmap *render_page_band(
	fz_context *ctx,
	fz_page *page,
	fz_irect bbox,
	float dpi,
	int alpha
) {
	fz_pixmap *pix = NULL;
	fz_device *dev = NULL;
	fz_matrix ctm = fz_scale(dpi / 72.0f, dpi / 72.0f);

	fz_try(ctx) {
		pix = fz_new_pixmap_with_bbox(ctx, fz_device_rgb(ctx), bbox, NULL, alpha);
		fz_clear_pixmap_with_value(ctx, pix, 255);

		dev = fz_new_draw_device(ctx, ctm, pix);
		fz_enable_device_hints(ctx, dev, FZ_NO_CACHE);
		fz_run_page(ctx, page, dev, fz_identity, NULL);
		fz_close_device(ctx, dev);
	}
	fz_always(ctx) {
		fz_drop_device(ctx, dev);
	}
	fz_catch(ctx) {
		fz_drop_pixmap(ctx, pix);
		return NULL;
	}

	return pix;
}

static fz_display_list *make_page_display_list(fz_context *ctx, fz_page *page) {
	fz_display_list *list = NULL;
	fz_device *dev = NULL;

	fz_try(ctx) {
		list = fz_new_display_list(ctx, fz_bound_page(ctx, page));
		dev = fz_new_list_device(ctx, list);
		fz_enable_device_hints(ctx, dev, FZ_NO_CACHE);
		fz_run_page(ctx, page, dev, fz_identity, NULL);
		fz_close_device(ctx, dev);
	}
	fz_always(ctx) {
		fz_drop_device(ctx, dev);
	}
	fz_catch(ctx) {
		fz_drop_display_list(ctx, list);
		return NULL;
	}

	return list;
}

static int render_display_band(
	fz_context *ctx,
	fz_display_list *list,
	fz_pixmap *pix,
	fz_irect page_bbox,
	int band_offset,
	int band_height,
	float dpi,
	int alpha
) {
	fz_device *dev = NULL;
	fz_matrix ctm = fz_scale(dpi / 72.0f, dpi / 72.0f);
	fz_rect clip = fz_rect_from_irect(page_bbox);
	clip.y0 += band_offset;
	clip.y1 = clip.y0 + band_height;
	clip.y1 += 2;

	fz_try(ctx) {
		pix->y = page_bbox.y0 + band_offset;
		fz_clear_pixmap_with_value(ctx, pix, 255);
		dev = fz_new_draw_device_with_proof(ctx, fz_identity, pix, NULL);
		fz_enable_device_hints(ctx, dev, FZ_NO_CACHE);
		fz_run_display_list(ctx, list, dev, ctm, clip, NULL);
		fz_close_device(ctx, dev);
	}
	fz_always(ctx) {
		fz_drop_device(ctx, dev);
	}
	fz_catch(ctx) {
		return 0;
	}

	return 1;
}

static fz_pixmap *new_band_pixmap(
	fz_context *ctx,
	fz_irect page_bbox,
	int band_height,
	int alpha
) {
	fz_irect bbox = page_bbox;
	if (bbox.y0 + band_height < bbox.y1)
		bbox.y1 = bbox.y0 + band_height;
	return fz_new_pixmap_with_bbox(ctx, fz_device_rgb(ctx), bbox, NULL, alpha);
}

static void clear_render_store(fz_context *ctx) {
	fz_empty_store(ctx);
}
*/
import "C"

import (
	"fmt"
	"runtime"
	"unsafe"
)

// RenderPageBands renders one page without allocating a full-page Go image.
// The callback must consume each band before returning and must not retain it.
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

	f.mtx.Lock()
	defer f.mtx.Unlock()

	if pageNumber < 0 || pageNumber >= f.NumPage() {
		return ErrPageMissing
	}

	page := C.load_page_safe(f.ctx, f.doc, C.int(pageNumber))
	if page == nil {
		return ErrLoadPage
	}
	defer C.fz_drop_page(f.ctx, page)

	var pageBox C.fz_irect
	if C.band_page_bounds(f.ctx, page, C.float(options.DPI), &pageBox) == 0 {
		return fmt.Errorf("fitz: invalid rendered page bounds")
	}

	var displayList *C.fz_display_list
	if options.UseDisplayList {
		displayList = C.make_page_display_list(f.ctx, page)
		if displayList == nil {
			return ErrRunPageContents
		}
		defer C.fz_drop_display_list(f.ctx, displayList)
	}

	var reusablePixmap *C.fz_pixmap
	if displayList != nil {
		reusablePixmap = C.new_band_pixmap(
			f.ctx,
			pageBox,
			C.int(options.BandHeight),
			C.int(boolToInt(options.Alpha)),
		)
		if reusablePixmap == nil {
			return ErrCreatePixmap
		}
		defer C.fz_drop_pixmap(f.ctx, reusablePixmap)
	}

	totalHeight := int(pageBox.y1 - pageBox.y0)
	for offset := 0; offset < totalHeight; offset += options.BandHeight {
		bandHeight := options.BandHeight
		if remaining := totalHeight - offset; remaining < bandHeight {
			bandHeight = remaining
		}

		bandBox := pageBox
		bandBox.y0 = pageBox.y0 + C.int(offset)
		bandBox.y1 = bandBox.y0 + C.int(bandHeight)

		renderBox := bandBox

		var pix *C.fz_pixmap
		if displayList != nil {
			if C.render_display_band(
				f.ctx,
				displayList,
				reusablePixmap,
				pageBox,
				C.int(offset),
				C.int(bandHeight),
				C.float(options.DPI),
				C.int(boolToInt(options.Alpha)),
			) == 0 {
				return ErrRunPageContents
			}
			pix = reusablePixmap
		} else {
			pix = C.render_page_band(
				f.ctx,
				page,
				renderBox,
				C.float(options.DPI),
				C.int(boolToInt(options.Alpha)),
			)
		}
		if pix == nil {
			return ErrRunPageContents
		}

		width := int(C.fz_pixmap_width(f.ctx, pix))
		renderHeight := int(C.fz_pixmap_height(f.ctx, pix))
		stride := int(C.fz_pixmap_stride(f.ctx, pix))
		components := int(C.fz_pixmap_components(f.ctx, pix))
		samples := C.fz_pixmap_samples(f.ctx, pix)
		if samples == nil || width <= 0 || renderHeight <= 0 || stride < width*components {
			if displayList == nil {
				C.fz_drop_pixmap(f.ctx, pix)
			}
			return ErrPixmapSamples
		}

		renderData := C.GoBytes(
			unsafe.Pointer(samples),
			C.int(stride*renderHeight),
		)
		if displayList == nil {
			C.fz_drop_pixmap(f.ctx, pix)
		}
		C.clear_render_store(f.ctx)

		renderOffset := int(bandBox.y0 - renderBox.y0)
		data := make([]byte, stride*bandHeight)
		for row := 0; row < bandHeight; row++ {
			src := (renderOffset + row) * stride
			dst := row * stride
			copy(data[dst:dst+stride], renderData[src:src+stride])
		}

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
