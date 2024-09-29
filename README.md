## go-fitz
[![Build Status](https://github.com/gen2brain/go-fitz/actions/workflows/test.yml/badge.svg)](https://github.com/gen2brain/go-fitz/actions)
[![GoDoc](https://godoc.org/github.com/gen2brain/go-fitz?status.svg)](https://godoc.org/github.com/gen2brain/go-fitz)
[![Go Report Card](https://goreportcard.com/badge/github.com/gen2brain/go-fitz?branch=master)](https://goreportcard.com/report/github.com/gen2brain/go-fitz)

Go wrapper for [MuPDF](http://mupdf.com/) fitz library that can extract pages from PDF, EPUB, MOBI, DOCX, XLSX and PPTX documents as IMG, TXT, HTML or SVG.

### Build tags

* `extlib` - use external MuPDF library
* `static` - build with static external MuPDF library (used with `extlib`)
* `pkgconfig` - enable pkg-config (used with `extlib`)
* `musl` - use musl compiled library
* `nocgo` - experimental [purego](https://github.com/ebitengine/purego) implementation (can also be used with `CGO_ENABLED=0`)

### Notes

The bundled libraries are built without CJK fonts, if you need them you must use the external library.

Calling e.g. Image() or Text() methods concurrently for the same document is not supported.

Purego implementation requires `libffi` and `libmupdf` shared libraries on runtime.
You must set `fitz.FzVersion` in your code or set `FZ_VERSION` environment variable to exact version of the shared library. 
    
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
