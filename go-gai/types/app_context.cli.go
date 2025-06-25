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

import "github.com/spf13/cobra"

// GetFileFlags returns the values of `--file` and “
func (app *AppContext) GetFileFlags() ([]string, []string) {
	file := app.Files
	if len(file) == 0 {
		file = append(file, app.RCFile.Defaults.Flags.File...)
	}

	files := app.FilePatterns
	if len(files) == 0 {
		files = append(files, app.RCFile.Defaults.Flags.Files...)
	}

	return file, files
}

// WithChatCLIFlags sets up `cmd` for chat based CLI flags.
func (app *AppContext) WithChatCLIFlags(cmd *cobra.Command) {
	app.WithPromptCLIFlags(cmd)
}

// WithDryRunCliFlags sets up `cmd` for dry run based CLI flags.
func (app *AppContext) WithDryRunCliFlags(cmd *cobra.Command) {
	cmd.Flags().BoolVarP(&app.DryRun, "dry-run", "", false, "do a dry run")
}

// WithDatabaseCLIFlags sets up `cmd` for database based CLI flags.
func (app *AppContext) WithDatabaseCLIFlags(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&app.Database, "database", "", "", "uri or path to database")
}

// WithEditorCLIFlags sets up `cmd` for editor based CLI flags.
func (app *AppContext) WithEditorCLIFlags(cmd *cobra.Command) {
	cmd.Flags().BoolVarP(&app.OpenEditor, "edit", "", false, "open editor")
	cmd.Flags().StringVarP(&app.Editor, "editor", "", "", "custom editor command")
}

// WithHighlightCLIFlags sets up `cmd` for highlight based CLI flags.
func (app *AppContext) WithHighlightCLIFlags(cmd *cobra.Command) {
	cmd.Flags().BoolVarP(&app.NoHighlight, "no-highlight", "", false, "do not highlight output")
}

// WithLanguageCLIFlags sets up `cmd` for language based CLI flags.
func (app *AppContext) WithLanguageCLIFlags(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&app.OutputLanguage, "language", "", "", "custom output language")
}

// WithPromptCLIFlags sets up `cmd` for prompt based CLI flags.
func (app *AppContext) WithPromptCLIFlags(cmd *cobra.Command) {
	app.WithEditorCLIFlags(cmd)
	app.WithHighlightCLIFlags(cmd)
	app.WithSchemaCLIFlags(cmd)
}

// WithSchemaCLIFlags sets up `cmd` for (response) format based CLI flags.
func (app *AppContext) WithSchemaCLIFlags(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&app.SchemaFile, "schema", "", "", "file with response format/schema")
	cmd.Flags().StringVarP(&app.SchemaName, "schema-name", "", "", "name of the response format/schema")
}

// WithYesCliFlags sets up `cmd` for "yes" based CLI flags.
func (app *AppContext) WithYesCliFlags(cmd *cobra.Command) {
	cmd.Flags().BoolVarP(&app.AlwaysYes, "yes", "y", false, "always yes")
}
