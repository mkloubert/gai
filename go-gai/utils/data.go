// MIT License
//
// Copyright (c) 2025 Marcel Joachim Kloubert (https://marcel.coffee)
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package utils

import (
	"archive/zip"
	"bytes"
	"encoding/base64"
	"encoding/csv"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"net/http"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/ledongthuc/pdf"
	"github.com/microcosm-cc/bluemonday"
	"github.com/xuri/excelize/v2"

	"golang.org/x/image/bmp"
	"golang.org/x/image/tiff"
	"golang.org/x/image/webp"
)

// ImageDecode describes a function that reads data from
// an `io.Reader` and create an `image.Image` instance if possible.
type ImageDecode = func(r io.Reader) (image.Image, error)

// ImageEncode describes a function that encodes an `img`
// to a byte array if possible.
type ImageEncode = func(img image.Image) ([]byte, error)

// ConvertImage converts an image to a specific type.
func ConvertImage(data []byte, encode ImageEncode) ([]byte, error) {
	mimeType := DetectMime(data)

	var decode ImageDecode = nil

	if strings.HasSuffix(mimeType, "/jpeg") || strings.HasSuffix(mimeType, "/jpg") {
		decode = jpeg.Decode
	} else if strings.HasSuffix(mimeType, "/webp") {
		decode = webp.Decode
	} else if strings.HasSuffix(mimeType, "/png") {
		decode = png.Decode
	} else if strings.HasSuffix(mimeType, "/bmp") {
		decode = bmp.Decode
	} else if strings.HasSuffix(mimeType, "/gif") {
		decode = gif.Decode
	} else if strings.HasSuffix(mimeType, "/tiff") {
		decode = tiff.Decode
	}

	if decode != nil {
		img, err := ReadImageFromBuffer(decode, data)
		if err != nil {
			return data, err
		}

		return encode(img)
	}
	return data, fmt.Errorf("type '%s' is not supported", mimeType)
}

// DataURIToBytes converts `dataURI` to byte array.
func DataURIToBytes(dataURI string) ([]byte, error) {
	const base64Prefix = ";base64,"

	idx := strings.Index(dataURI, base64Prefix)
	if idx < 0 {
		return nil, fmt.Errorf("not a base64 data URI")
	}

	base64Data := dataURI[idx+len(base64Prefix):]
	data, err := base64.StdEncoding.DecodeString(base64Data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// DetectMime is an extension of `http.DetectContentType`.
func DetectMime(b []byte) string {
	if len(b) >= 12 {
		// HEIC & co.

		if bytes.Equal(b[4:8], []byte("ftyp")) {
			brand := string(b[8:12])
			switch brand {
			case "heic", "heix", "hevc", "hevx":
				return "image/heic"
			case "mif1", "msf1":
				return "image/heif"
			case "avif":
				return "image/avif"
			}
		}
	} else if len(b) >= 8 {
		// old Excel format

		header := b[0:8]
		magic := []byte{0xD0, 0xCF, 0x11, 0xE0, 0xA1, 0xB1, 0x1A, 0xE1}

		if bytes.Equal(header, magic) {
			return "application/vnd.ms-excel"
		}
	}

	isXLSXFile, err := IsXLSX(b)
	if err == nil {
		if isXLSXFile {
			return "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
		}
	}

	isPPTXFile, err := IsPPTX(b)
	if err == nil {
		if isPPTXFile {
			return "application/vnd.openxmlformats-officedocument.presentationml.presentation"
		}
	}

	isDOCXFile, err := IsDOCX(b)
	if err == nil {
		if isDOCXFile {
			return "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
		}
	}

	return http.DetectContentType(b)
}

// EnsurePlainText keeps sure that input `data`
// becomes plain text.
func EnsurePlainText(data []byte) (string, error) {
	mimeType := DetectMime(data)

	if strings.HasSuffix(mimeType, "/vnd.openxmlformats-officedocument.presentationml.presentation") {
		// PowerPoint PPTX

		r := bytes.NewReader(data)

		size := int64(len(data))

		z, err := zip.NewReader(r, size)
		if err != nil {
			return "", err
		}

		texts := make([]string, 0)
		getJoinedText := func() string {
			return strings.Join(texts, "\n\n\n")
		}

		for _, f := range z.File {
			if !(strings.HasPrefix(f.Name, "ppt/slides/slide") && strings.HasSuffix(f.Name, ".xml")) {
				continue
			}

			rc, err := f.Open()
			if err != nil {
				continue
			}
			rc.Close()

			buf := new(bytes.Buffer)
			_, err = io.Copy(buf, rc)
			if err != nil {
				continue
			}

			text := &strings.Builder{}

			// extract all <a:t>...</a:t>
			decoder := xml.NewDecoder(bytes.NewReader(buf.Bytes()))
			for {
				t, err := decoder.Token()
				if err == io.EOF {
					break
				}
				if err != nil {
					break
				}

				switch se := t.(type) {
				case xml.StartElement:
					if se.Name.Local == "t" {
						// Text-Tag gefunden
						var innerText string
						decoder.DecodeElement(&innerText, &se)

						text.WriteString(innerText)
						text.WriteString("\n")
					}
				}
			}

			texts = append(texts, text.String())
		}

		return getJoinedText(), nil
	} else if strings.HasSuffix(mimeType, "/vnd.openxmlformats-officedocument.wordprocessingml.document") {
		// Word DOCX

		r := bytes.NewReader(data)

		size := int64(len(data))

		z, err := zip.NewReader(r, size)
		if err != nil {
			return "", err
		}

		var docXMLFile *zip.File

		for _, f := range z.File {
			if f.Name == "word/document.xml" {
				docXMLFile = f
				break
			}
		}

		texts := make([]string, 0)
		getJoinedText := func() string {
			return strings.Join(texts, "\n\n\n")
		}

		if docXMLFile != nil {
			rc, err := docXMLFile.Open()
			if err == nil {
				// XML could be opened
				defer rc.Close()

				buf := &bytes.Buffer{}
				_, err = io.Copy(buf, rc)
				if err == nil {
					// data from XML could be copied
					decoder := xml.NewDecoder(buf)

					// collect text
					var text strings.Builder
					for {
						t, err := decoder.Token()
						if err == io.EOF {
							break // ignore errors
						}
						if err != nil {
							break // ignore errors
						}

						switch se := t.(type) {
						case xml.CharData:
							// only the text
							text.WriteString(string(se))
						}
					}

					texts = append(texts, text.String())
				}
			}
		}

		return getJoinedText(), nil
	} else if strings.HasSuffix(mimeType, "/vnd.openxmlformats-officedocument.spreadsheetml.sheet") || strings.HasSuffix(mimeType, "/vnd.ms-excel") {
		// Excel

		r := bytes.NewReader(data)

		f, err := excelize.OpenReader(r)
		if err != nil {
			return "", err
		}

		defer f.Close()

		texts := make([]string, 0)
		getJoinedText := func() string {
			return strings.Join(texts, "\n\n\n")
		}

		sheets := f.GetSheetList()
		for _, s := range sheets {
			buff := &bytes.Buffer{}

			sheetName := s

			jsonData, err := json.Marshal(&sheetName)
			if err == nil {
				sheetName = string(jsonData)
			} else {
				sheetName = fmt.Sprintf(`"%s"`, sheetName)
			}

			rows, err := f.GetRows(s)
			if err != nil {
				return getJoinedText(), err
			}

			writer := csv.NewWriter(buff)

			for _, record := range rows {
				err := writer.Write(record)
				if err != nil {
					return getJoinedText(), err
				}
			}

			writer.Flush()

			err = writer.Error()
			if err != nil {
				return getJoinedText(), err
			}

			texts = append(texts, buff.String())
		}

		return getJoinedText(), nil
	} else if strings.HasSuffix(mimeType, "/htm") || strings.HasSuffix(mimeType, "/html") {
		// HTML

		r := bytes.NewReader(data)

		doc, err := goquery.NewDocumentFromReader(r)
		if err == nil {
			var sel *goquery.Selection

			body := doc.Has("body")
			if body != nil {
				sel = body.Contents()
			} else {
				sel = doc.Contents()
			}

			if sel != nil {
				sanitized := bluemonday.UGCPolicy().Sanitize(sel.Text())

				return strings.TrimSpace(sanitized), nil
			}
		}
	} else if strings.HasSuffix(mimeType, "/pdf") {
		// PDF

		r := bytes.NewReader(data)

		size := int64(len(data))

		pdf, err := pdf.NewReader(r, size)
		if err != nil {
			return "", err
		}

		b, err := pdf.GetPlainText()
		if err != nil {
			return "", err
		}

		var buf bytes.Buffer
		defer buf.Reset()

		_, err = buf.ReadFrom(b)
		if err != nil {
			return "", err
		}
		return buf.String(), err
	}

	return string(data), nil
}

// EnsurePNG ensures having a image in PNG format.
func EnsurePNG(data []byte) ([]byte, error) {
	mimeType := DetectMime(data)

	if strings.HasSuffix(mimeType, "/png") {
		return data, nil
	}

	encodeImage := func(img image.Image) ([]byte, error) {
		writer := &bytes.Buffer{}

		err := png.Encode(writer, img)

		return writer.Bytes(), err
	}

	return ConvertImage(data, encodeImage)
}

// GetPartsOfDataURI converts returns the parts of `dataURI`.
func GetPartsOfDataURI(dataURI string) (string, string, error) {
	parts := strings.SplitN(dataURI, ",", 2)
	if len(parts) != 2 {
		return "", "", errors.New("invalid data URI")
	}

	meta := strings.TrimPrefix(parts[0], "data:")
	base64Data := parts[1]

	mimeParts := strings.SplitN(meta, ";", 2)
	if len(mimeParts) < 1 {
		return base64Data, "", fmt.Errorf("no MIME type found")
	}

	return base64Data, strings.TrimSpace(
		strings.ToLower(mimeParts[0]),
	), nil
}

// MaybeBinary checks if data maybe binary / non-printable.
func MaybeBinary(data []byte) bool {
	// check for non-printable chars except whitespaces
	for _, b := range data {
		if (b == 0) || (b < 7 || (b > 13 && b < 32)) {
			return true // maybe binary
		}
	}

	return false
}

// IsDOCX checks if `data` contains a Word file in DOCX format.
func IsDOCX(data []byte) (bool, error) {
	return IsOfficeFile(data, "word")
}

// IsOfficeFile checks if `data` contains is file in Office format.
func IsOfficeFile(data []byte, folderName string) (bool, error) {
	r := bytes.NewReader(data)

	size := int64(len(data))

	z, err := zip.NewReader(r, size)
	if err != nil {
		return false, err
	}

	foundContentTypes := false
	foundFolder := false

	for _, f := range z.File {
		fullFolderName := fmt.Sprintf("%s/", folderName)

		// need this file
		if f.Name == "[Content_Types].xml" {
			foundContentTypes = true
		}
		// need a file in `folderName`
		if len(f.Name) > len(fullFolderName) && strings.HasPrefix(f.Name, fullFolderName) {
			foundFolder = true
		}

		if foundContentTypes && foundFolder {
			return true, nil // seems to be a Office file
		}
	}

	return false, nil
}

// IsPPTX checks if `data` contains a PowerPoint file in PPTX format.
func IsPPTX(data []byte) (bool, error) {
	return IsOfficeFile(data, "ppt")
}

// IsXLSX checks if `data` contains a Excel file in XLSX format.
func IsXLSX(data []byte) (bool, error) {
	return IsOfficeFile(data, "xl")
}

// ReadImageFromBuffer reads an `image.Image` instance from byte array with a `types.ImageDecode` function.
func ReadImageFromBuffer(decode ImageDecode, data []byte) (image.Image, error) {
	reader := bytes.NewReader(data)

	return webp.Decode(reader)
}
