go-fitz
========

Simple Golang wrapper for the [MuPDF](http://mupdf.com/) Fitz library that can extract images from PDF, EPUB and XPS documents.

Install
-------

MuPDF version 1.8 is required:

    $ git clone git://git.ghostscript.com/mupdf.git && cd mupdf
    $ git submodule update --init --recursive
    $ curl -L https://gist.githubusercontent.com/gen2brain/7869ac4c6db5933f670f/raw/1619394dc957ae10bcd73c713760993466b4bfea/mupdf-openssl-curl.patch | patch -p1
    $ sed -e "1iHAVE_X11 = no" -e "1iWANT_OPENSSL = no" -e "1iWANT_CURL = no" -i Makerules
    $ HAVE_X11=no HAVE_GLFW=no HAVE_GLUT=no WANT_OPENSSL=no WANT_CURL=no HAVE_MUJS=yes HAVE_JSCORE=no HAVE_V8=no make && make install

    $ go get github.com/gen2brain/go-fitz

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
