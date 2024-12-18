package fitz_test

import (
	"errors"
	"fmt"
	"image"
	"image/jpeg"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/gen2brain/go-fitz"
)

func TestImage(t *testing.T) {
	doc, err := fitz.New(filepath.Join("testdata", "test.pdf"))
	if err != nil {
		t.Error(err)
	}

	defer doc.Close()

	tmpDir, err := os.MkdirTemp(os.TempDir(), "fitz")
	if err != nil {
		t.Error(err)
	}

	defer os.RemoveAll(tmpDir)

	for n := 0; n < doc.NumPage(); n++ {
		img, err := doc.Image(n)
		if err != nil {
			t.Error(err)
		}

		f, err := os.Create(filepath.Join(tmpDir, fmt.Sprintf("test%03d.jpg", n)))
		if err != nil {
			t.Error(err)
		}

		err = jpeg.Encode(f, img, &jpeg.Options{Quality: jpeg.DefaultQuality})
		if err != nil {
			t.Error(err)
		}

		f.Close()
	}
}

func TestImageFromMemory(t *testing.T) {
	b, err := os.ReadFile(filepath.Join("testdata", "test.pdf"))
	if err != nil {
		t.Error(err)
	}

	doc, err := fitz.NewFromMemory(b)
	if err != nil {
		t.Error(err)
	}

	defer doc.Close()

	tmpDir, err := os.MkdirTemp(os.TempDir(), "fitz")
	if err != nil {
		t.Error(err)
	}

	defer os.RemoveAll(tmpDir)

	for n := 0; n < doc.NumPage(); n++ {
		img, err := doc.Image(n)
		if err != nil {
			t.Error(err)
		}

		f, err := os.Create(filepath.Join(tmpDir, fmt.Sprintf("test%03d.jpg", n)))
		if err != nil {
			t.Error(err)
		}

		err = jpeg.Encode(f, img, &jpeg.Options{Quality: jpeg.DefaultQuality})
		if err != nil {
			t.Error(err)
		}

		f.Close()
	}
}

func TestLinks(t *testing.T) {
	doc, err := fitz.New(filepath.Join("testdata", "test.pdf"))
	if err != nil {
		t.Error(err)
	}

	defer doc.Close()

	links, err := doc.Links(2)
	if err != nil {
		t.Error(err)
	}

	if len(links) != 1 {
		t.Error("expected 1 link, got", len(links))
	}

	if links[0].URI != "https://creativecommons.org/licenses/by-nc-sa/4.0/" {
		t.Error("expected empty URI, got", links[0].URI)
	}
}

func TestText(t *testing.T) {
	doc, err := fitz.New(filepath.Join("testdata", "test.pdf"))
	if err != nil {
		t.Error(err)
	}

	defer doc.Close()

	tmpDir, err := os.MkdirTemp(os.TempDir(), "fitz")
	if err != nil {
		t.Error(err)
	}

	defer os.RemoveAll(tmpDir)

	for n := 0; n < doc.NumPage(); n++ {
		text, err := doc.Text(n)
		if err != nil {
			t.Error(err)
		}

		f, err := os.Create(filepath.Join(tmpDir, fmt.Sprintf("test%03d.txt", n)))
		if err != nil {
			t.Error(err)
		}

		_, err = f.WriteString(text)
		if err != nil {
			t.Error(err)
		}

		f.Close()
	}
}

func TestHTML(t *testing.T) {
	doc, err := fitz.New(filepath.Join("testdata", "test.pdf"))
	if err != nil {
		t.Error(err)
	}

	defer doc.Close()

	tmpDir, err := os.MkdirTemp(os.TempDir(), "fitz")
	if err != nil {
		t.Error(err)
	}

	defer os.RemoveAll(tmpDir)

	for n := 0; n < doc.NumPage(); n++ {
		html, err := doc.HTML(n, true)
		if err != nil {
			t.Error(err)
		}

		f, err := os.Create(filepath.Join(tmpDir, fmt.Sprintf("test%03d.html", n)))
		if err != nil {
			t.Error(err)
		}

		_, err = f.WriteString(html)
		if err != nil {
			t.Error(err)
		}

		f.Close()
	}
}

func TestPNG(t *testing.T) {
	doc, err := fitz.New(filepath.Join("testdata", "test.pdf"))
	if err != nil {
		t.Error(err)
	}

	defer doc.Close()

	tmpDir, err := os.MkdirTemp(os.TempDir(), "fitz")
	if err != nil {
		t.Error(err)
	}

	defer os.RemoveAll(tmpDir)

	for n := 0; n < doc.NumPage(); n++ {
		png, err := doc.ImagePNG(n, 300.0)
		if err != nil {
			t.Error(err)
		}

		if err = os.WriteFile(filepath.Join(tmpDir, fmt.Sprintf("test%03d.png", n)), png, 0644); err != nil {
			t.Error(err)
		}
	}
}

func TestSVG(t *testing.T) {
	doc, err := fitz.New(filepath.Join("testdata", "test.pdf"))
	if err != nil {
		t.Error(err)
	}

	defer doc.Close()

	tmpDir, err := os.MkdirTemp(os.TempDir(), "fitz")
	if err != nil {
		t.Error(err)
	}

	defer os.RemoveAll(tmpDir)

	for n := 0; n < doc.NumPage(); n++ {
		svg, err := doc.SVG(n)
		if err != nil {
			t.Error(err)
		}

		f, err := os.Create(filepath.Join(tmpDir, fmt.Sprintf("test%03d.svg", n)))
		if err != nil {
			t.Error(err)
		}

		_, err = f.WriteString(svg)
		if err != nil {
			t.Error(err)
		}

		f.Close()
	}
}

func TestToC(t *testing.T) {
	doc, err := fitz.New(filepath.Join("testdata", "test.pdf"))
	if err != nil {
		t.Error(err)
	}

	defer doc.Close()

	_, err = doc.ToC()
	if err != nil {
		t.Error(err)
	}
}

func TestMetadata(t *testing.T) {
	doc, err := fitz.New(filepath.Join("testdata", "test.pdf"))
	if err != nil {
		t.Error(err)
	}

	defer doc.Close()

	meta := doc.Metadata()
	if len(meta) == 0 {
		t.Error(fmt.Errorf("metadata is empty"))
	}
}

func TestBound(t *testing.T) {
	doc, err := fitz.New(filepath.Join("testdata", "test.pdf"))
	if err != nil {
		t.Error(err)
	}

	defer doc.Close()
	expected := image.Rect(0, 0, 612, 792)

	for i := 0; i < doc.NumPage(); i++ {
		bound, err := doc.Bound(i)
		if err != nil {
			t.Error(err)
		}
		if bound != expected {
			t.Error(fmt.Errorf("bounds didn't match go %v when expedient %v", bound, expected))
		}
	}

	_, err = doc.Bound(doc.NumPage())
	if !errors.Is(err, fitz.ErrPageMissing) {
		t.Error(fmt.Errorf("ErrPageMissing not returned got %v", err))
	}
}

func TestEmptyBytes(t *testing.T) {
	var err error
	// empty reader
	_, err = fitz.NewFromReader(emptyReader{})
	if !errors.Is(err, fitz.ErrEmptyBytes) {
		t.Errorf("Expected ErrEmptyBytes, got %v", err)
	}
	// nil slice
	_, err = fitz.NewFromMemory(nil)
	if !errors.Is(err, fitz.ErrEmptyBytes) {
		t.Errorf("Expected ErrEmptyBytes, got %v", err)
	}
	// empty slice
	_, err = fitz.NewFromMemory(make([]byte, 0))
	if !errors.Is(err, fitz.ErrEmptyBytes) {
		t.Errorf("Expected ErrEmptyBytes, got %v", err)
	}
}

type emptyReader struct{}

func (emptyReader) Read([]byte) (int, error) { return 0, io.EOF }
