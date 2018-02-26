## go-fitz

Go wrapper for [MuPDF](http://mupdf.com/) fitz library that can extract images from PDF, EPUB and XPS documents.

### Install

    go get -u github.com/gen2brain/go-fitz

### Example

    doc, err := fitz.New("test.pdf")
    if err != nil {
        panic(err)
    }

    defer doc.Close()

    for n := 0; n < doc.Pages(); n++ {
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
