package fitz_test

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"image/jpeg"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"testing"
	"time"

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

func TestConcurrent(t *testing.T) {
	for _, name := range []string{"test.pdf", "bug_jpx.pdf"} {
		b, err := os.ReadFile(filepath.Join("testdata", name))
		if err != nil {
			t.Fatal(err)
		}

		const workers = 8
		const iters = 20

		var wg sync.WaitGroup
		errs := make(chan error, workers*iters)

		for w := 0; w < workers; w++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for i := 0; i < iters; i++ {
					doc, err := fitz.NewFromMemory(b)
					if err != nil {
						errs <- err
						return
					}
					if _, err := doc.ImageDPI(0, 100); err != nil {
						doc.Close()
						errs <- err
						return
					}
					if _, err := doc.Text(0); err != nil {
						doc.Close()
						errs <- err
						return
					}
					doc.Close()
				}
			}()
		}

		wg.Wait()
		close(errs)
		for err := range errs {
			t.Errorf("%s: %v", name, err)
		}
	}
}

func TestAnnotations(t *testing.T) {
	doc, err := fitz.New(filepath.Join("testdata", "annot.pdf"))
	if err != nil {
		t.Fatal(err)
	}
	defer doc.Close()

	img, err := doc.ImageDPI(0, 72)
	if err != nil {
		t.Fatal(err)
	}

	if c := img.RGBAAt(100, 100); c.R != 255 || c.G != 0 || c.B != 0 {
		t.Errorf("annotation not rendered, center pixel %v", c)
	}

	if c := img.RGBAAt(10, 10); c.R != 255 || c.G != 255 || c.B != 255 {
		t.Errorf("unexpected corner pixel %v", c)
	}
}

func TestCloseTwice(t *testing.T) {
	doc, err := fitz.New(filepath.Join("testdata", "test.pdf"))
	if err != nil {
		t.Fatal(err)
	}

	doc.Close()
	doc.Close()
}

func openDropped(t *testing.T) *fitz.Document {
	b, err := os.ReadFile(filepath.Join("testdata", "test.pdf"))
	if err != nil {
		t.Fatal(err)
	}

	doc, err := fitz.NewFromMemory(b)
	if err != nil {
		t.Fatal(err)
	}

	return doc
}

func TestMemoryOwnership(t *testing.T) {
	doc := openDropped(t)
	defer doc.Close()

	for i := 0; i < 5; i++ {
		runtime.GC()
	}

	if _, err := doc.ImageDPI(0, 50); err != nil {
		t.Error(err)
	}
}

func TestNoClose(t *testing.T) {
	for i := 0; i < 5; i++ {
		openDropped(t)
		runtime.GC()
		time.Sleep(10 * time.Millisecond)
	}
}

func imagePDF(t *testing.T, pageSize, imageSize int) []byte {
	var jpg bytes.Buffer
	if err := jpeg.Encode(&jpg, image.NewRGBA(image.Rect(0, 0, 2000, 2000)), nil); err != nil {
		t.Fatal(err)
	}

	content := fmt.Sprintf("q %d 0 0 %d 0 0 cm /Im0 Do Q", imageSize, imageSize)
	objs := []string{
		"<< /Type /Catalog /Pages 2 0 R >>",
		"<< /Type /Pages /Kids [3 0 R] /Count 1 >>",
		fmt.Sprintf("<< /Type /Page /Parent 2 0 R /MediaBox [0 0 %d %d] /Resources << /XObject << /Im0 5 0 R >> >> /Contents 4 0 R >>", pageSize, pageSize),
		fmt.Sprintf("<< /Length %d >>\nstream\n%s\nendstream", len(content), content),
		fmt.Sprintf("<< /Type /XObject /Subtype /Image /Width 2000 /Height 2000 /ColorSpace /DeviceRGB /BitsPerComponent 8 /Filter /DCTDecode /Length %d >>\nstream\n%s\nendstream", jpg.Len(), jpg.String()),
	}

	var pdf bytes.Buffer
	pdf.WriteString("%PDF-1.5\n")
	offsets := make([]int, len(objs))
	for i, obj := range objs {
		offsets[i] = pdf.Len()
		fmt.Fprintf(&pdf, "%d 0 obj\n%s\nendobj\n", i+1, obj)
	}

	xref := pdf.Len()
	fmt.Fprintf(&pdf, "xref\n0 %d\n0000000000 65535 f \n", len(objs)+1)
	for _, off := range offsets {
		fmt.Fprintf(&pdf, "%010d 00000 n \n", off)
	}
	fmt.Fprintf(&pdf, "trailer\n<< /Size %d /Root 1 0 R >>\nstartxref\n%d\n%%%%EOF\n", len(objs)+1, xref)

	return pdf.Bytes()
}

func TestImageNativeDPI(t *testing.T) {
	for _, tt := range []struct {
		name              string
		pageSize, imgSize int
		width, height     int
	}{
		{"small image", 612, 20, 2550, 2550},
		{"full page image", 200, 200, 2000, 2000},
	} {
		t.Run(tt.name, func(t *testing.T) {
			doc, err := fitz.NewFromMemory(imagePDF(t, tt.pageSize, tt.imgSize))
			if err != nil {
				t.Fatal(err)
			}
			defer doc.Close()

			img, err := doc.Image(0)
			if err != nil {
				t.Fatal(err)
			}

			if b := img.Bounds(); b.Dx() != tt.width || b.Dy() != tt.height {
				t.Errorf("got %dx%d, want %dx%d", b.Dx(), b.Dy(), tt.width, tt.height)
			}
		})
	}
}
