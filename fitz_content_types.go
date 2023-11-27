package fitz

import "bytes"

// contentType returns document MIME type.
func contentType(b []byte) string {
	l := len(b)
	// for file length shortcuts see https://github.com/mathiasbynens/small
	switch {
	case l < 8:
		return ""
	case isPAM(b):
		return "image/x-portable-arbitrarymap"
	case isPBM(b):
		return "image/x-portable-bitmap"
	case isPFM(b):
		return "image/x-portable-floatmap"
	case isPGM(b):
		return "image/x-portable-greymap"
	case isPPM(b):
		return "image/x-portable-pixmap"
	case isGIF(b):
		return "image/gif"
	case l < 16:
		return ""
	case isBMP(b):
		return "image/bmp"
	case isJBIG2(b):
		// file header + segment header = 24 bytes
		return "image/x-jb2"
	case l < 32:
		return ""
	case isTIFF(b):
		return "image/tiff"
	case isSVG(b):
		// min of 41 bytes: <svg xmlns="http://www.w3.org/2000/svg"/>
		return "image/svg+xml"
	case l < 64:
		return ""
	case isJPEG(b):
		return "image/jpeg"
	case isPNG(b):
		return "image/png"
	case isJPEG2000(b):
		return "image/jp2"
	case isJPEGXR(b):
		return "image/vnd.ms-photo"
	case isPDF(b):
		return "application/pdf"
	case isPSD(b):
		return "image/vnd.adobe.photoshop"
	case isZIP(b):
		switch {
		case isEPUB(b):
			return "application/epub+zip"
		case isXPS(b):
			return "application/oxps"
		default:
			// fitz will consider it a Comic Book Archive
			// must contain at least one image, i.e. >64 bytes
			return "application/zip"
		}
	case isXML(b):
		// fitz will consider it an FB2
		// minimal valid FB2 w/o content is >64 bytes
		return "text/xml"
	case isMOBI(b):
		return "application/x-mobipocket-ebook"
	default:
		return ""
	}
}

func isBMP(b []byte) bool {
	return b[0] == 0x42 && b[1] == 0x4D
}

func isGIF(b []byte) bool {
	return b[0] == 0x47 && b[1] == 0x49 && b[2] == 0x46 && b[3] == 0x38
}

func isJBIG2(b []byte) bool {
	return b[0] == 0x97 && b[1] == 0x4A && b[2] == 0x42 && b[3] == 0x32 &&
		b[4] == 0x0D && b[5] == 0x0A && b[6] == 0x1A && b[7] == 0x0A
}

func isJPEG(b []byte) bool {
	return b[0] == 0xFF && b[1] == 0xD8 && b[2] == 0xFF
}

func isJPEG2000(b []byte) bool {
	switch {
	case b[0] == 0xFF && b[1] == 0x4F && b[2] == 0xFF && b[3] == 0x51:
		return true
	default:
		return b[0] == 0x00 && b[1] == 0x00 && b[2] == 0x00 && b[3] == 0x0C &&
			b[4] == 0x6A && b[5] == 0x50 && b[6] == 0x20 && b[7] == 0x20 &&
			b[8] == 0x0D && b[9] == 0x0A && b[10] == 0x87 && b[11] == 0x0A
	}
}

func isJPEGXR(b []byte) bool {
	return b[0] == 0x49 && b[1] == 0x49 && b[2] == 0xBC
}

func isPAM(b []byte) bool {
	return b[0] == 0x50 && b[1] == 0x37 && b[2] == 0x0A
}

func isPBM(b []byte) bool {
	return b[0] == 0x50 && (b[1] == 0x31 || b[1] == 0x34) && b[2] == 0x0A
}

func isPFM(b []byte) bool {
	return b[0] == 0x50 && (b[1] == 0x46 || b[1] == 0x66) && b[2] == 0x0A
}

func isPGM(b []byte) bool {
	return b[0] == 0x50 && (b[1] == 0x32 || b[1] == 0x35) && b[2] == 0x0A
}

func isPPM(b []byte) bool {
	return b[0] == 0x50 && (b[1] == 0x33 || b[1] == 0x36) && b[2] == 0x0A
}

func isPNG(b []byte) bool {
	return b[0] == 0x89 && b[1] == 0x50 && b[2] == 0x4E && b[3] == 0x47 &&
		b[4] == 0x0D && b[5] == 0x0A && b[6] == 0x1A && b[7] == 0x0A
}

func isTIFF(b []byte) bool {
	return b[0] == 0x49 && b[1] == 0x49 && b[2] == 0x2A && b[3] == 0x00 ||
		b[0] == 0x4D && b[1] == 0x4D && b[2] == 0x00 && b[3] == 0x2A
}

// PDF magic number 25 50 44 46 = "%PDF".
func isPDF(b []byte) bool {
	return b[0] == 0x25 && b[1] == 0x50 && b[2] == 0x44 && b[3] == 0x46
}

// PSD magic number 38 42 50 53 = "8BPS"
func isPSD(b []byte) bool {
	return b[0] == 0x38 && b[1] == 0x42 && b[2] == 0x50 && b[3] == 0x53
}

// Non-empty ZIP archive magic number 50 4B 03 04.
func isZIP(b []byte) bool {
	return b[0] == 0x50 && b[1] == 0x4B && b[2] == 0x03 && b[3] == 0x04
}

// Looks for a file named "mimetype" containing the ASCII string "application/epub+zip".
// The file must be uncompressed and be the first file within the archive.
func isEPUB(b []byte) bool {
	return b[30] == 0x6D && b[31] == 0x69 && b[32] == 0x6D && b[33] == 0x65 &&
		b[34] == 0x74 && b[35] == 0x79 && b[36] == 0x70 && b[37] == 0x65 &&
		b[38] == 0x61 && b[39] == 0x70 && b[40] == 0x70 && b[41] == 0x6C &&
		b[42] == 0x69 && b[43] == 0x63 && b[44] == 0x61 && b[45] == 0x74 &&
		b[46] == 0x69 && b[47] == 0x6F && b[48] == 0x6E && b[49] == 0x2F &&
		b[50] == 0x65 && b[51] == 0x70 && b[52] == 0x75 && b[53] == 0x62 &&
		b[54] == 0x2B && b[55] == 0x7A && b[56] == 0x69 && b[57] == 0x70
}

// MOBI contains either BOOKMOBI or TEXtREAd string after a 60 bytes offset.
// The magic string is then followed by at least 10 bytes of information.
func isMOBI(b []byte) bool {
	switch {
	case len(b) < 78:
		return false
	case b[60] == 0x42 && b[61] == 0x4F && b[62] == 0x4F && b[63] == 0x4B &&
		b[64] == 0x4D && b[65] == 0x4F && b[66] == 0x42 && b[67] == 0x49:
		return true
	case b[60] == 0x54 && b[61] == 0x45 && b[62] == 0x58 && b[63] == 0x74 &&
		b[64] == 0x52 && b[65] == 0x45 && b[66] == 0x41 && b[67] == 0x64:
		return true
	default:
		return false
	}
}

// Looks for a file named "[Content_Types].xml" at the root of a ZIP archive.
// MS Office apps put this file first within the archive enabling for fast detection.
func isXPS(b []byte) bool {
	return b[30] == 0x5B && b[31] == 0x43 && b[32] == 0x6F && b[33] == 0x6E &&
		b[34] == 0x74 && b[35] == 0x65 && b[36] == 0x6E && b[37] == 0x74 &&
		b[38] == 0x5F && b[39] == 0x54 && b[40] == 0x79 && b[41] == 0x70 &&
		b[42] == 0x65 && b[43] == 0x73 && b[44] == 0x5D && b[45] == 0x2E &&
		b[46] == 0x78 && b[47] == 0x6D && b[48] == 0x6C
}

// Checks for "<svg" string after XML prolog, DOCTYPE and comments.
// See svg_recognize_doc_content in mupdf/source/svg/svg-doc.c
func isSVG(b []byte) bool {
	if b[0] == 0xEF && b[1] == 0xBB {
		b = b[2:] // ignore UTF-8 BOM
	}
	r := bytes.NewReader(b)
ParseSVGText:
	for {
		if c, err := r.ReadByte(); err == nil {
			switch c {
			case 0x09, 0x0A, 0x0D, 0x20: // whitespace
				continue
			case 0x3C: // <
				goto ParseSVGElement
			default:
				return false
			}
		}
		return false
	}
ParseSVGElement:
	if c, err := r.ReadByte(); err != nil {
		return false
	} else if c == 0x21 || c == 0x3F { // ! or ?
		goto ParseSVGComment
	} else if c != 0x73 { // s
		return false
	} else if c, err := r.ReadByte(); err != nil || c != 0x76 { // v
		return false
	} else if c, err := r.ReadByte(); err != nil || c != 0x67 { // g
		return false
	}
	return true
ParseSVGComment:
	for {
		c, err := r.ReadByte()
		if err != nil {
			return false
		} else if c == 0x3E { // >
			goto ParseSVGText
		}
	}
}

// Checks for "<?xml" string at the beginning of the file.
// Possible occurrences of a UTF-8 BOM are also considered.
func isXML(b []byte) bool {
	switch {
	// w/o UTF-8 BOM:
	case b[0] == 0x3C && b[1] == 0x3F && b[2] == 0x78 && b[3] == 0x6D && b[4] == 0x6C:
		return true
	// w/ UTF-8 BOM:
	default:
		return b[0] == 0xEF && b[1] == 0xBB && b[2] == 0xBF && b[3] == 0x3C &&
			b[4] == 0x3F && b[5] == 0x78 && b[6] == 0x6D && b[7] == 0x6C
	}
}
