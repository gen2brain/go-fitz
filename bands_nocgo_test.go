//go:build !cgo || nocgo

package fitz

import "testing"

func TestRenderPageBandsNocgo(t *testing.T) {
	doc, err := New("testdata/test.pdf")
	if err != nil {
		t.Fatal(err)
	}
	defer doc.Close()

	bands := 0
	rows := 0
	err = doc.RenderPageBands(0, BandOptions{
		DPI:        72,
		BandHeight: 64,
		Alpha:      true,
	}, func(band Band) error {
		if band.Width <= 0 || band.Height <= 0 || band.Components != 4 {
			t.Fatalf("invalid band: %+v", band)
		}
		if band.Height > 64 || band.Y != rows {
			t.Fatalf("invalid band position or height: y=%d height=%d", band.Y, band.Height)
		}
		rows += band.Height
		bands++
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if bands < 2 || rows <= 0 {
		t.Fatalf("unexpected band output: bands=%d rows=%d", bands, rows)
	}
}
