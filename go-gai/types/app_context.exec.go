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
	"os"
	"os/exec"
	"runtime"
	"strings"
)

// CreateExecCommand creates a new pre-setuped command.
func (app *AppContext) CreateExecCommand(f string, args ...string) *exec.Cmd {
	cmd := exec.Command(f, args...)

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd
}

// RunExecCommand runs a new pre-setuped command.
func (app *AppContext) RunExecCommand(f string, args ...string) (*exec.Cmd, error) {
	cmd := app.CreateExecCommand(f, args...)

	cmd.Dir = app.WorkingDirectory

	err := cmd.Run()

	return cmd, err
}

// TryGetBestOpenEditorCommand tries to find and return the best command to open a file for editing with the given file path.
// It returns the command and its arguments as a slice of strings.
// If no suitable editor is found, it returns an empty string and an empty slice.
func (app *AppContext) TryGetBestOpenEditorCommand(filePath string) (string, []string) {
	osName := runtime.GOOS

	// first check for custom editor set in CLI
	customEditor := strings.TrimSpace(app.Editor)
	if customEditor == "" {
		// ... then in environment variable
		customEditor = strings.TrimSpace(app.GetEnv("GAI_EDITOR"))
	}
	if customEditor != "" {
		return app.TryGetExecutablePath(customEditor), []string{filePath}
	}

	// check for editors based on OS
	if osName == "windows" {
		return "notepad.exe", []string{filePath}
	}

	// try vi editor
	viPath := app.TryGetExecutablePath("vi")
	if viPath != "" {
		return viPath, []string{"-c", "startinsert", filePath}
	}

	// try nano editor
	nanoPath := app.TryGetExecutablePath("nano")
	if nanoPath != "" {
		return nanoPath, []string{filePath}
	}

	// if no editor was found, return empty strings
	return "", []string{}
}

// TryGetExecutablePath tries to find and return the path of the given command executable file.
// It returns the path as a string. If the executable file is not found, it returns an empty string.
func (app *AppContext) TryGetExecutablePath(command string) string {
	exePath, err := exec.LookPath(command)
	if err != nil {
		return ""
	}

	return exePath
}
