## go-fitz
[![GoDoc](https://godoc.org/github.com/gen2brain/go-fitz?status.svg)](https://godoc.org/github.com/gen2brain/go-fitz)

Go wrapper for [MuPDF](http://mupdf.com/) fitz library 
that can extract pages from PDF, EPUB and XPS documents as images or text.

### Install

    go get -u github.com/gen2brain/go-fitz

### Example
```go
doc, err := fitz.New("test.pdf")
if err != nil {
    panic(err)
}

defer doc.Close()

// Extract pages as images
for n := 0; n < doc.NumPage(); n++ {
    img, err := doc.Image(n)
    if err != nil {
        panic(err)
    }

    f, err := os.Create(fmt.Sprintf("test%03d.jpg", n))
    if err != nil {
        panic(err)
    }

    err = jpeg.Encode(f, img, &jpeg.Options{jpeg.DefaultQuality})
    if err != nil {
        panic(err)
    }

    f.Close()
}

// Extract pages as text
for n := 0; n < doc.NumPage(); n++ {
    img, err := doc.Text(n)
    if err != nil {
        panic(err)
    }

    f, err := os.Create(fmt.Sprintf("test%03d.txt", n))
    if err != nil {
        panic(err)
    }

    f.Close()
}

```
