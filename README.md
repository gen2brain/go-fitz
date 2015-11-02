go-fitz
========

Simple Golang wrapper for the [MuPDF](http://mupdf.com/) Fitz library that can extract images from documents.

Install
-------

MuPDF version 1.8 is required.

    go get github.com/gen2brain/go-fitz

Example
-------

	doc, err := fitz.NewDocument("test.pdf")
	if err != nil {
		fmt.Printf("Error: %s\n", err.Error())
	}
	defer doc.Close()

	for n := 0; n < doc.Pages(); n++ {
		img, _ := doc.Image(n)
		f, _ := os.Create(fmt.Sprintf("test%03d.jpg", n))
		jpeg.Encode(f, img, &jpeg.Options{jpeg.DefaultQuality})
		f.Close()
	}
