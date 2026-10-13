package fitz_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/gen2brain/go-fitz"
)

// Run with -cpu=1,2,4,8 to measure rendering across independent documents.
func BenchmarkConcurrentImageDPI(b *testing.B) {
	for _, name := range []string{"test.pdf", "bug_jpx.pdf"} {
		b.Run(name, func(b *testing.B) {
			data, err := os.ReadFile(filepath.Join("testdata", name))
			if err != nil {
				b.Fatal(err)
			}

			b.ReportAllocs()
			b.ResetTimer()
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					doc, err := fitz.NewFromMemory(data)
					if err != nil {
						b.Error(err)
						return
					}
					_, renderErr := doc.ImageDPI(0, 200)
					closeErr := doc.Close()
					if renderErr != nil {
						b.Error(renderErr)
						return
					}
					if closeErr != nil {
						b.Error(closeErr)
						return
					}
				}
			})
		})
	}
}
