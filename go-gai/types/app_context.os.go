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

package types

import (
	"fmt"
	"os"
	"path/filepath"
)

// FindDirUp tries to find a directory by `n` by walking
// step-by-step up from `WorkingDirectory` until it reaches the root level.
func (app *AppContext) FindDirUp(n string) (string, error) {
	dir := app.WorkingDirectory
	for {
		gitPath := filepath.Join(dir, n)

		info, err := os.Stat(gitPath)

		if err == nil && info.IsDir() {
			return gitPath, nil
		} else if err != nil {
			if !os.IsNotExist(err) {
				return dir, err
			}
		}

		parentDir := filepath.Dir(dir)
		if parentDir == dir {
			break // root directory reached
		}

		dir = parentDir
	}

	return dir, fmt.Errorf("directory '%s' not found", n)
}
