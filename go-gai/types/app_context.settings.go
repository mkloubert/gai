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
	"strings"

	"github.com/goccy/go-yaml"
)

func (app *AppContext) loadRCFile() {
	rcFile := &GAIRCFile{}

	possibleRCFiles := make([]string, 0)
	possibleRCFiles = append(possibleRCFiles, filepath.Join(app.WorkingDirectory, ".gairc"))
	possibleRCFiles = append(possibleRCFiles, filepath.Join(app.WorkingDirectory, ".gairc.yaml"))
	possibleRCFiles = append(possibleRCFiles, filepath.Join(app.WorkingDirectory, ".gairc.yml"))

	existingRCFiles := make([]string, 0)
	for _, f := range possibleRCFiles {
		stat, err := os.Stat(f)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}

			app.CheckIfError(err)
		}

		if !stat.IsDir() {
			existingRCFiles = append(existingRCFiles, f)
		}
	}

	if len(existingRCFiles) > 1 {
		// 0 or 1, not more

		app.CheckIfError(fmt.Errorf("there are more than 1 possible files: %s", strings.Join(existingRCFiles, ",")))
	} else if len(existingRCFiles) == 1 {
		data, err := os.ReadFile(existingRCFiles[0])
		app.CheckIfError(err)

		err = yaml.Unmarshal(data, rcFile)
		app.CheckIfError(err)
	}

	app.RCFile = rcFile
}
