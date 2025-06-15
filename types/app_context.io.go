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
	"runtime"
	"strings"
)

func (app *AppContext) getBestChromaFormatterName() string {
	terminalFormatter := strings.TrimSpace(app.TerminalFormatter)
	if terminalFormatter == "" {
		terminalFormatter = strings.TrimSpace(
			app.Getenv("GAI_TERMINAL_FORMATTER"),
		)
	}

	if terminalFormatter != "" {
		return terminalFormatter
	}

	switch os := runtime.GOOS; os {
	case "darwin", "linux":
		return "terminal16m"
	case "windows":
		return "terminal256"
	}

	return "terminal"
}

func (app *AppContext) getBestChromaStyleName() string {
	terminalStyle := strings.TrimSpace(app.TerminalStyle)
	if terminalStyle == "" {
		terminalStyle = strings.TrimSpace(
			app.Getenv("GAI_TERMINAL_STYLE"),
		)
	}

	if terminalStyle != "" {
		return terminalStyle
	}

	return "dracula"
}

// GetChromaSettings returns an instance of `ChromaSettings`.
func (app *AppContext) GetChromaSettings() *ChromaSettings {
	return &ChromaSettings{
		App:       app,
		Formatter: app.getBestChromaFormatterName(),
		Style:     app.getBestChromaStyleName(),
	}
}

// Write writes `b` to `Stdout`.
func (app *AppContext) Write(b []byte) (n int, err error) {
	return app.Stdout.Write(b)
}

// Writeln writes `s` separated by space to `Stdout` and adds `EOL`.
func (app *AppContext) Writeln(vals ...any) (n int, err error) {
	s := make([]string, 0)
	for _, v := range vals {
		s = append(s, fmt.Sprintf("%v", v))
	}

	return app.WriteString(fmt.Sprintf("%v%v", strings.Join(s, " "), app.EOL))
}

// WriteError writes `b` to `Stderr`.
func (app *AppContext) WriteError(b []byte) (n int, err error) {
	return app.Stderr.Write(b)
}

// WriteErrorString writes `s` to `Stderr`.
func (app *AppContext) WriteErrorString(s string) (n int, err error) {
	return app.Stderr.WriteString(s)
}

// WriteString writes `s` to `Stdout`.
func (app *AppContext) WriteString(s string) (n int, err error) {
	return app.Stdout.WriteString(s)
}
