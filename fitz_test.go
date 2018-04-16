package fitz

import (
	"fmt"
	"image/jpeg"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestImage(t *testing.T) {
	doc, err := New(filepath.Join("testdata", "test.pdf"))
	if err != nil {
		t.Error(err)
	}

	defer doc.Close()

	tmpDir, err := ioutil.TempDir(os.TempDir(), "fitz")
	if err != nil {
		t.Error(err)
	}

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
	b, err := ioutil.ReadFile(filepath.Join("testdata", "test.pdf"))
	if err != nil {
		t.Error(err)
	}

	doc, err := NewFromMemory(b)
	if err != nil {
		t.Error(err)
	}

	defer doc.Close()

	tmpDir, err := ioutil.TempDir(os.TempDir(), "fitz")
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

func TestText(t *testing.T) {
	doc, err := New(filepath.Join("testdata", "test.pdf"))
	if err != nil {
		t.Error(err)
	}

	defer doc.Close()

	tmpDir, err := ioutil.TempDir(os.TempDir(), "fitz")
	if err != nil {
		t.Error(err)
	}

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
	doc, err := New(filepath.Join("testdata", "test.pdf"))
	if err != nil {
		t.Error(err)
	}

	defer doc.Close()

	tmpDir, err := ioutil.TempDir(os.TempDir(), "fitz")
	if err != nil {
		t.Error(err)
	}

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
