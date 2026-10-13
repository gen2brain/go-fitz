//go:build cgo && !nocgo

package fitz_test

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/gen2brain/go-fitz"
)

func TestOpenErrors(t *testing.T) {
	for _, tc := range []struct {
		name      string
		data      []byte
		memoryErr error
	}{
		{"unsupported", []byte("not a document"), fitz.ErrOpenMemory},
		{"broken.pdf", append([]byte("%PDF-1.7\n"), bytes.Repeat([]byte("invalid\n"), 10)...), fitz.ErrOpenDocument},
	} {
		t.Run(tc.name, func(t *testing.T) {
			filename := filepath.Join(t.TempDir(), tc.name)
			if err := os.WriteFile(filename, tc.data, 0600); err != nil {
				t.Fatal(err)
			}

			doc, err := fitz.New(filename)
			if doc != nil {
				doc.Close()
				t.Error("New returned a document after an open failure")
			}
			if !errors.Is(err, fitz.ErrOpenDocument) {
				t.Errorf("New error = %v, want %v", err, fitz.ErrOpenDocument)
			}

			doc, err = fitz.NewFromMemory(tc.data)
			if doc != nil {
				doc.Close()
				t.Error("NewFromMemory returned a document after an open failure")
			}
			if !errors.Is(err, tc.memoryErr) {
				t.Errorf("NewFromMemory error = %v, want %v", err, tc.memoryErr)
			}
		})
	}

	doc, err := fitz.New(filepath.Join(t.TempDir(), "missing.pdf"))
	if doc != nil || !errors.Is(err, fitz.ErrNoSuchFile) {
		t.Fatalf("New(missing file) = (%v, %v), want (nil, %v)", doc, err, fitz.ErrNoSuchFile)
	}
}
