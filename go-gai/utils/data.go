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
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"net/http"
	"strings"

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
	// HEIC & co.
	if len(b) >= 12 {
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
	}

	return http.DetectContentType(b)
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

// ReadImageFromBuffer reads an `image.Image` instance from byte array with a `types.ImageDecode` function.
func ReadImageFromBuffer(decode ImageDecode, data []byte) (image.Image, error) {
	reader := bytes.NewReader(data)

	return webp.Decode(reader)
}
