//go:build !cgo || nocgo

package fitz

// RenderPageBands is unavailable when MuPDF is loaded without the cgo backend.
func (f *Document) RenderPageBands(
	pageNumber int,
	options BandOptions,
	consume func(Band) error,
) error {
	return ErrBandedRenderingUnavailable
}
