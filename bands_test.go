//go:build cgo && !nocgo

package fitz

import (
	"testing"
)

func TestRenderPageBands(t *testing.T) {
	doc, err := New("testdata/test.pdf")
	if err != nil {
		t.Fatal(err)
	}
	defer doc.Close()

	width := 0
	height := 0
	bands := 0
	err = doc.RenderPageBands(0, BandOptions{
		DPI:            72,
		BandHeight:     64,
		Alpha:          true,
		UseDisplayList: true,
	}, func(band Band) error {
		if band.Width <= 0 || band.Height <= 0 || band.Components != 4 {
			t.Fatalf("invalid band: %+v", band)
		}
		if band.Stride < band.Width*band.Components || len(band.Samples) < band.Stride*band.Height {
			t.Fatalf("invalid band samples: %+v", band)
		}
		if band.Height > 64 {
			t.Fatalf("band exceeds requested height: %d", band.Height)
		}
		if band.Y != height {
			t.Fatalf("non-contiguous band: got y=%d, want %d", band.Y, height)
		}
		width = band.Width
		height += band.Height
		bands++
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if bands < 2 {
		t.Fatalf("expected multiple bands, got %d", bands)
	}
	if width <= 0 || height <= 0 {
		t.Fatalf("invalid rendered dimensions: %dx%d", width, height)
	}
}

func TestRenderPageBandsValidation(t *testing.T) {
	doc := &Document{}
	consumer := func(Band) error { return nil }

	for name, options := range map[string]BandOptions{
		"invalid dpi":         {DPI: 0, BandHeight: 64},
		"invalid band height": {DPI: 72, BandHeight: 0},
	} {
		t.Run(name, func(t *testing.T) {
			if err := doc.RenderPageBands(0, options, consumer); err == nil {
				t.Fatal("expected validation error")
			}
		})
	}
	if err := doc.RenderPageBands(0, BandOptions{DPI: 72, BandHeight: 64}, nil); err == nil {
		t.Fatal("expected nil consumer error")
	}
}
