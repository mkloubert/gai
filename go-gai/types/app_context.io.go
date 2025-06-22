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
	"bufio"
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
)

func (app *AppContext) getBestChromaFormatterName() string {
	terminalFormatter := strings.TrimSpace(app.TerminalFormatter)
	if terminalFormatter == "" {
		terminalFormatter = strings.TrimSpace(
			app.GetEnv("GAI_TERMINAL_FORMATTER"),
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
			app.GetEnv("GAI_TERMINAL_STYLE"),
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

// GetInput retrieves user input from the command-line arguments, standard input, or an editor.
func (app *AppContext) GetInput(args []string) (string, error) {
	stdin := app.Stdin

	GAI_INPUT_ORDER := app.GetEnv("GAI_INPUT_ORDER")

	GAI_INPUT_SEPARATOR := app.GetEnvOrNil("GAI_INPUT_SEPARATOR")
	if GAI_INPUT_SEPARATOR == nil {
		defaultSep := " "

		GAI_INPUT_SEPARATOR = &defaultSep
	}

	// first add arguments from CLI
	var parts []string

	// addPart function trims the white spaces and appends the non-empty strings to the parts slice
	addPart := func(val string) {
		val = strings.TrimSpace(val)
		if val != "" {
			parts = append(parts, val)
		}
	}

	// add arguments passed in the command-line interface
	readFromArgs := func() error {
		addPart(strings.Join(args, " "))
		return nil
	}

	// if `OpenEditor` is true, attempts to open the user's preferred text editor and waits for the user to input text.
	readFromEditor := func() error {
		if !app.OpenEditor {
			return nil
		}

		// create a temporary file
		tmpFile, err := os.CreateTemp("", "gai")
		if err != nil {
			return err
		}

		tmpFile.Close()

		tmpFilePath, err := filepath.Abs(tmpFile.Name())
		if err != nil {
			return err
		}

		// get the command to open the editor
		editorPath, editorArgs := app.TryGetBestOpenEditorCommand(tmpFilePath)
		if editorPath == "" {
			return errors.New("no matching editor found")
		}

		defer os.Remove(tmpFilePath)

		// run the editor command
		cmd := app.CreateExecCommand(editorPath, editorArgs...)
		cmd.Dir = path.Dir(tmpFilePath)

		err = cmd.Run()
		if err != nil {
			return err
		}

		// read the contents of the temporary file
		tmpData, err := os.ReadFile(tmpFilePath)
		if err != nil {
			return err
		}

		addPart(string(tmpData))
		return nil
	}

	var dataFromStdin *string
	dataFromStdinChecked := false
	readFromStdin := func() error {
		if !dataFromStdinChecked {
			dataFromStdinChecked = true

			// check if standard input has been piped
			stdinStat, _ := stdin.Stat()
			if (stdinStat.Mode() & os.ModeCharDevice) == 0 {
				scanner := bufio.NewScanner(stdin)

				temp := ""
				for scanner.Scan() {
					temp += scanner.Text()
				}

				dataFromStdin = &temp
			}
		}

		if dataFromStdin != nil {
			addPart(*dataFromStdin)
		}

		return nil
	}

	// collect actions
	inputActions := make([](func() error), 0)
	for _, item := range strings.Split(GAI_INPUT_ORDER, ",") {
		item = strings.TrimSpace(strings.ToLower(item))
		if item == "" {
			continue
		}

		switch item {
		case "a", "args":
			inputActions = append(inputActions, readFromArgs)
		case "e", "editor":
			inputActions = append(inputActions, readFromEditor)
		case "in", "stdin":
			inputActions = append(inputActions, readFromStdin)
		default:
			return "", fmt.Errorf("input order '%v' not supported", item)
		}
	}

	if len(inputActions) == 0 {
		// setup default
		inputActions = append(inputActions, readFromArgs, readFromStdin, readFromEditor)
	}

	// invoke actions
	for _, action := range inputActions {
		err := action()
		if err != nil {
			return "", err
		}
	}

	return strings.TrimSpace(
		strings.Join(parts, *GAI_INPUT_SEPARATOR),
	), nil
}

// GetOutputFile returns the path to the file where to write output to
func (app *AppContext) GetOutputFile() string {
	outputFile := strings.TrimSpace(app.OutputFile) // first try flags
	if outputFile == "" {
		outputFile = strings.TrimSpace(app.GetEnv("GAI_OUTPUT_FILE")) // now try env var
	}

	if outputFile != "" {
		// ensure its absolute
		if !filepath.IsAbs(outputFile) {
			outputFile = filepath.Join(app.WorkingDirectory, outputFile)
		}
	}

	return outputFile
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
