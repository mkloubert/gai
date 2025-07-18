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
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"
	"time"

	"github.com/mkloubert/gai/utils"
	ignore "github.com/sabhiram/go-gitignore"
	"github.com/spf13/cobra"
)

// AppContext handles data and logic for running application.
type AppContext struct {
	// AI is the default AI client.
	AI AIClient
	// AlwaysYes is `true` if command should answer any user interactions with "yes".
	AlwaysYes bool
	// ApiKey stores a global API key.
	ApiKey string
	// BaseUrl stores base URL.
	BaseUrl string
	// CommandPath stores full path of current command.
	CommandPath []string
	// Context stores the name of the current context.
	Context string
	// Database stores the path or URI to the database, usually a SQLite database.
	Database string
	// DryRun is `true` if command should be run in "dry run mode".
	DryRun bool
	// Editor stores the command for the custom editor to use.
	Editor string
	// EnvFiles stores list of additional .env files that should be loaded in this direction.
	EnvFiles []string
	// EnvFiles stores environment variables.
	EnvVars map[string]string
	// EnvFiles stores string representing new line.
	EOL string
	// FilePatterns stores list of additional files as glob patterns to use for the current operation.
	FilePatterns []string
	// Files stores list of additional files to use for the current operation.
	Files []string
	// HomeDirectory is the absolute path to the user's home directory.
	HomeDirectory string
	// Log is the logger the app should use.
	Log *log.Logger
	// MaxTokens stores the maximum number of tokens.
	MaxTokens int64
	// Model is the default chat model to use.
	Model string
	// NoHighlight is `true` if output should NOT be highlighted and formatted.
	NoHighlight bool
	// OpenEditor is `true` if editor should be opened.
	OpenEditor bool
	// OutputFile stores where to store the ouput of the app to.
	OutputFile string
	// OutputLanguage stores the output language.
	OutputLanguage string
	// RCFile stores current `.gairc` file.
	RCFile *GAIRCFile
	// RootCommand stores the root command.
	RootCommand *cobra.Command
	// SchemaFile stores the path to the file with the response format/schema.
	SchemaFile string
	// SchemaFile stores the name of the response format/schema.
	SchemaName string
	// SkipDefaultEnvFiles indicates not to use default .env files, if `true`.
	SkipDefaultEnvFiles bool
	// Stderr stores the stream for error outputs.
	Stderr *os.File
	// Stdin stores the stream for default inputs.
	Stdin *os.File
	// Stdout stores the stream for default outputs.
	Stdout *os.File
	// SystemPrompt stores the custom system prompt for AI operations.
	SystemPrompt string
	// SystemRole custom name of the system role.
	SystemRole string
	// TempDirectory stores the path of the custom temp directory.
	TempDirectory string
	// Temperature stores the temperature for AI operations.
	Temperature float64
	// TerminalFormatter defines the custom terminal formatter.
	TerminalFormatter string
	// TerminalFormatter defines the custom terminal style.
	TerminalStyle string
	// Verbose indicates if application should also output debug messages.
	Verbose bool
	// WorkingDirectory stores the current root directory.
	WorkingDirectory string
}

// CheckIfError checks if `err` is not `nil` and exists in this case.
func (app *AppContext) CheckIfError(err error) {
	if err != nil {
		app.WriteErrorString(fmt.Sprintf("%s%s", err.Error(), app.EOL))
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

// GetCurrentContext returns the name of the current AI context.
func (app *AppContext) GetCurrentContext() string {
	context := strings.TrimSpace(app.Context)
	if context != "" {
		return context
	}

	return strings.TrimSpace(app.GetEnv("GAI_CONTEXT"))
}

// GetFiles() returns the cleaned up and unsorted list of files as
// full paths from `Files`.
func (app *AppContext) GetFiles() ([]string, error) {
	fileFlag, filesFlag := app.GetFileFlags()

	files := make([]string, 0)

	// first check for explicit file pathes
	for _, f := range fileFlag {
		file := f
		if !filepath.IsAbs(file) {
			file = filepath.Join(app.WorkingDirectory, file)
		}

		files = append(files, file)
	}

	// now the ones with patterns ...

	globPatterns := make([]string, 0)
	for _, fp := range filesFlag {
		if strings.TrimSpace(fp) != "" {
			globPatterns = append(globPatterns, fp)
		}
	}
	globPatterns = utils.RemoveDuplicateStrings(globPatterns)

	if len(globPatterns) > 0 {
		gitignore := ignore.CompileIgnoreLines(globPatterns...)

		err := filepath.WalkDir(app.WorkingDirectory, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}

			if !d.IsDir() {
				relPath, err := filepath.Rel(app.WorkingDirectory, path)
				if err != nil {
					return err
				}

				if gitignore.MatchesPath(relPath) {
					files = append(files, path)
				}
			}

			return nil
		})

		if err != nil {
			return files, err
		}
	}

	// remove duplicates ...
	files = utils.RemoveDuplicateStrings(files)

	// ... and sort case-insensitive
	sort.Slice(files, func(i, j int) bool {
		strI := strings.TrimSpace(strings.ToLower(files[i]))
		strJ := strings.TrimSpace(strings.ToLower(files[j]))

		return strI < strJ
	})

	return files, nil
}

// GetFullPath returns full path of `p`.
func (app *AppContext) GetFullPath(p string) string {
	fullPath := strings.TrimSpace(p)
	if fullPath == "" {
		return app.WorkingDirectory
	}

	if !filepath.IsAbs(fullPath) {
		fullPath = filepath.Join(app.WorkingDirectory, fullPath)
	}

	return fullPath
}

// GetISOTime() returns current timestamp in UTC ISO 8601 format.
func (app *AppContext) GetISOTime() string {
	return app.GetNow().Format("2006-01-02T15:04:05.000Z")
}

// GetNow() returns current timestamp in UTC format.
func (app *AppContext) GetNow() time.Time {
	return time.Now().UTC()
}

// NewFilePredicate creates a new function that checks if a file path matches a specific pattern.
func (app *AppContext) NewFilePredicate() func(f string) (bool, error) {
	fileFlag, filesFlag := app.GetFileFlags()

	explicitFiles := make([]string, 0)

	// first check for explicit file pathes
	for _, f := range fileFlag {
		file := f
		if !filepath.IsAbs(file) {
			file = filepath.Join(app.WorkingDirectory, file)
		}

		explicitFiles = append(explicitFiles, file)
	}

	globPatterns := make([]string, 0)
	for _, fp := range filesFlag {
		if strings.TrimSpace(fp) != "" {
			globPatterns = append(globPatterns, fp)
		}
	}
	globPatterns = utils.RemoveDuplicateStrings(globPatterns)

	if len(explicitFiles) == 0 && len(globPatterns) == 0 {
		// nothing defined => allow all

		return func(f string) (bool, error) {
			return true, nil
		}
	}

	gitignore := ignore.CompileIgnoreLines(globPatterns...)

	return func(f string) (bool, error) {
		absPath := f
		if !filepath.IsAbs(absPath) {
			absPath = filepath.Join(app.WorkingDirectory)
		}

		if slices.Contains(explicitFiles, absPath) {
			return true, nil // explicit file found
		}

		relPath, err := filepath.Rel(app.WorkingDirectory, f)
		if err != nil {
			return false, err
		}

		// check glob patterns in .gitignore format
		return gitignore.MatchesPath(relPath), nil
	}
}

// Run runs the application.
func (app *AppContext) Run() {
	app.CheckIfError(
		app.RootCommand.Execute(),
	)
}
