package fitz

import (
	_ "embed"
	"testing"
)

func testContentType(want string, b []byte, t *testing.T) {
	if got := contentType(b); got != want {
		t.Errorf("contentType([]byte) = '%v'; want '%v'", got, want)
	}
}

//go:embed testdata/test.bmp
var bmp []byte

func TestContentTypeBMP(t *testing.T) {
	testContentType("image/bmp", bmp, t)
}

//go:embed testdata/test.epub
var epub []byte

func TestContentTypeEPUB(t *testing.T) {
	testContentType("application/epub+zip", epub, t)
}

//go:embed testdata/test.mobi
var mobi []byte

func TestContentTypeMOBI(t *testing.T) {
	testContentType("application/x-mobipocket-ebook", mobi, t)
}

//go:embed testdata/test.cbz
var cbz []byte

func TestContentTypeCBZ(t *testing.T) {
	testContentType("application/zip", cbz, t)
}

//go:embed testdata/test.fb2
var fb2 []byte

func TestContentTypeFB2(t *testing.T) {
	testContentType("text/xml", fb2, t)
}

//go:embed testdata/test.gif
var gif []byte

func TestContentTypeGIF(t *testing.T) {
	testContentType("image/gif", gif, t)
}

//go:embed testdata/test.jb2
var jb2 []byte

func TestContentTypeJBIG2(t *testing.T) {
	testContentType("image/x-jb2", jb2, t)
}

//go:embed testdata/test.jpg
var jpg []byte

func TestContentTypeJPEG(t *testing.T) {
	testContentType("image/jpeg", jpg, t)
}

//go:embed testdata/test.jp2
var jp2 []byte

func TestContentTypeJPEG2000(t *testing.T) {
	testContentType("image/jp2", jp2, t)
}

//go:embed testdata/test.jxr
var jxr []byte

func TestContentTypeJPEGXR(t *testing.T) {
	testContentType("image/vnd.ms-photo", jxr, t)
}

//go:embed testdata/test.pam
var pam []byte

func TestContentTypePAM(t *testing.T) {
	testContentType("image/x-portable-arbitrarymap", pam, t)
}

//go:embed testdata/test.pbm
var pbm []byte

func TestContentTypePBM(t *testing.T) {
	testContentType("image/x-portable-bitmap", pbm, t)
}

//go:embed testdata/test.pdf
var pdf []byte

func TestContentTypePDF(t *testing.T) {
	testContentType("application/pdf", pdf, t)
}

//go:embed testdata/test.psd
var psd []byte

func TestContentTypePSD(t *testing.T) {
	testContentType("image/vnd.adobe.photoshop", psd, t)
}

//go:embed testdata/test.pfm
var pfm []byte

func TestContentTypePFM(t *testing.T) {
	testContentType("image/x-portable-floatmap", pfm, t)
}

//go:embed testdata/test.pgm
var pgm []byte

func TestContentTypePGM(t *testing.T) {
	testContentType("image/x-portable-greymap", pgm, t)
}

//go:embed testdata/test.ppm
var ppm []byte

func TestContentTypePPM(t *testing.T) {
	testContentType("image/x-portable-pixmap", ppm, t)
}

//go:embed testdata/test.svg
var svg []byte

func TestContentTypeSVG(t *testing.T) {
	testContentType("image/svg+xml", svg, t)
}

//go:embed testdata/test.tif
var tif []byte

func TestContentTypeTIFF(t *testing.T) {
	testContentType("image/tiff", tif, t)
}

//go:embed testdata/test.xps
var xps []byte

func TestContentTypeXPS(t *testing.T) {
	testContentType("application/oxps", xps, t)
}
