package fitz

import "errors"

// BandOptions controls bounded page rendering.
type BandOptions struct {
	DPI            float64
	BandHeight     int
	Alpha          bool
	UseDisplayList bool
}

// Band contains one rendered horizontal page band.
// Samples is owned by the caller and must not be retained after the callback returns.
type Band struct {
	Y          int
	Width      int
	Height     int
	Stride     int
	Components int
	Samples    []byte
}

var ErrBandedRenderingUnavailable = errors.New("fitz: banded rendering unavailable")
