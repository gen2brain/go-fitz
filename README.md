## go-fitz
[![Build Status](https://github.com/gen2brain/go-fitz/actions/workflows/test.yml/badge.svg)](https://github.com/gen2brain/go-fitz/actions)
[![Go Reference](https://pkg.go.dev/badge/github.com/gen2brain/go-fitz.svg)](https://pkg.go.dev/github.com/gen2brain/go-fitz)

Go wrapper for [MuPDF](http://mupdf.com/) fitz library that can extract pages from PDF, EPUB, MOBI, DOCX, XLSX and PPTX documents as IMG, TXT, HTML or SVG.

For a pure Go alternative, see [folio](https://github.com/gen2brain/folio), document rendering in pure Go with no CGo.

### Build tags

* `extlib` - use external MuPDF library
* `static` - build with static external MuPDF library (used with `extlib`)
* `pkgconfig` - enable pkg-config (used with `extlib`)
* `musl` - use musl compiled library
* `nocgo` - experimental [purego](https://github.com/ebitengine/purego) implementation (can also be used with `CGO_ENABLED=0`)

### Notes

The bundled libraries are built without CJK fonts, if you need them you must use the external library.

Concurrent rendering is supported when each goroutine uses its own `Document` (a separate MuPDF context). Concurrency on a single `Document`, including racing with `Close`, is not supported.
With the `nocgo` implementation, calls into MuPDF are serialized across all documents, use the default cgo implementation for parallel rendering.

The purego implementation loads the `libmupdf` shared library at runtime. Its version is detected automatically; set the `FZ_VERSION` environment variable (or `fitz.FzVersion`) to override it.
    
### Example
```go
package main

import (
	"fmt"
	"image/jpeg"
	"os"
	"path/filepath"

	"github.com/gen2brain/go-fitz"
)

func main() {
	doc, err := fitz.New("test.pdf")
	if err != nil {
		panic(err)
	}

	defer doc.Close()

	tmpDir, err := os.MkdirTemp(os.TempDir(), "fitz")
	if err != nil {
		panic(err)
	}

	// Extract pages as images
	for n := 0; n < doc.NumPage(); n++ {
		img, err := doc.Image(n)
		if err != nil {
			panic(err)
		}

		f, err := os.Create(filepath.Join(tmpDir, fmt.Sprintf("test%03d.jpg", n)))
		if err != nil {
			panic(err)
		}

		err = jpeg.Encode(f, img, &jpeg.Options{jpeg.DefaultQuality})
		if err != nil {
			panic(err)
		}

		f.Close()
	}
}
```
