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
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
)

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
