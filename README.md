go-fitz
========

Simple Golang wrapper for the [MuPDF](http://mupdf.com/) Fitz library.


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
