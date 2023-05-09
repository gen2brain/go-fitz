package fitz_test

import (
	"fmt"
	"image/jpeg"
	"os"
	"path/filepath"

	"github.com/gen2brain/go-fitz"
)

func ExampleNew() {
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

		err = jpeg.Encode(f, img, &jpeg.Options{Quality: jpeg.DefaultQuality})
		if err != nil {
			panic(err)
		}

		f.Close()
	}

	// Extract pages as text
	for n := 0; n < doc.NumPage(); n++ {
		text, err := doc.Text(n)
		if err != nil {
			panic(err)
		}

		f, err := os.Create(filepath.Join(tmpDir, fmt.Sprintf("test%03d.txt", n)))
		if err != nil {
			panic(err)
		}

		_, err = f.WriteString(text)
		if err != nil {
			panic(err)
		}

		f.Close()
	}

	// Extract pages as html
	for n := 0; n < doc.NumPage(); n++ {
		html, err := doc.HTML(n, true)
		if err != nil {
			panic(err)
		}

		f, err := os.Create(filepath.Join(tmpDir, fmt.Sprintf("test%03d.html", n)))
		if err != nil {
			panic(err)
		}

		_, err = f.WriteString(html)
		if err != nil {
			panic(err)
		}

		f.Close()
	}

	// Extract pages as svg
	for n := 0; n < doc.NumPage(); n++ {
		svg, err := doc.SVG(n)
		if err != nil {
			panic(err)
		}

		f, err := os.Create(filepath.Join(tmpDir, fmt.Sprintf("test%03d.svg", n)))
		if err != nil {
			panic(err)
		}

		_, err = f.WriteString(svg)
		if err != nil {
			panic(err)
		}

		f.Close()
	}
}
