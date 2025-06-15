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
	"log"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

// AppContext handles data and logic for running application.
type AppContext struct {
	// AI is the default AI client.
	AI AIClient
	// ApiKey stores a global API key.
	ApiKey string
	// EnvFiles stores list of additional .env files that should be loaded in this direction.
	EnvFiles []string
	// EnvFiles stores environment variables.
	EnvVars map[string]string
	// EnvFiles stores string representing new line.
	EOL string
	// HomeDirectory is the absolute path to the user's home directory.
	HomeDirectory string
	// Log is the logger the app should use.
	Log *log.Logger
	// Model is the default chat model to use.
	Model string
	// RootCommand stores the root command.
	RootCommand *cobra.Command
	// SkipDefaultEnvFiles indicates not to use default .env files, if `true`.
	SkipDefaultEnvFiles bool
	// Stderr stores the stream for error outputs.
	Stderr *os.File
	// Stdin stores the stream for default inputs.
	Stdin *os.File
	// Stdout stores the stream for default outputs.
	Stdout *os.File
	// Verbose indicates if application should also output debug messages.
	Verbose bool
	// WorkingDirectory stores the current root directory.
	WorkingDirectory string
}

// CheckIfError checks if `err` is not `nil` and exists in this case.
func (app *AppContext) CheckIfError(err error) {
	if err != nil {
		fmt.Fprintln(app.Stderr, err)
		os.Exit(1)
	}
}

// Dbg outputs a `v` if app has "verbose flag" set and additionally adds a new line.
func (app *AppContext) Dbg(v any) {
	app.Dbgf("%v%v", v, app.EOL)
}

// Dbgf outputs a `v` if app has "verbose flag" set, similar to Printf().
func (app *AppContext) Dbgf(format string, v ...any) {
	if app.Verbose {
		app.Log.Printf(format, v...)
	}
}

// EnsureAppDir ensures that the root directory for this app inside
// `HomeDirectory` exists and returns its path.
func (app *AppContext) EnsureAppDir() (string, error) {
	appDir := filepath.Join(app.HomeDirectory, ".gai")

	if stat, err := os.Stat(appDir); err == nil {
		// exists

		if !stat.IsDir() {
			// ... but no file
			return appDir, fmt.Errorf("'%v' is not directory", appDir)
		}

		return appDir, nil
	} else if os.IsNotExist(err) {
		// does not exist => create
		err := os.MkdirAll(appDir, 0755)

		return appDir, err
	} else {
		// other error
		return appDir, err
	}
}

// Run runs the application.
func (app *AppContext) Run() {
	app.CheckIfError(
		app.RootCommand.Execute(),
	)
}
