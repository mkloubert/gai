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

// WithChatFlags sets up `cmd` for chat based CLI flags.
func (app *AppContext) WithChatFlags(cmd *cobra.Command) {
	app.WithPromptFlags(cmd)
}

// WithEditorCLIFlags sets up `cmd` for editor based CLI flags.
func (app *AppContext) WithEditorCLIFlags(cmd *cobra.Command) {
	cmd.Flags().BoolVarP(&app.OpenEditor, "edit", "", false, "open editor")
	cmd.Flags().StringVarP(&app.Editor, "editor", "", "", "custom editor command")
}

// WithHighlightFlags sets up `cmd` for highlight based CLI flags.
func (app *AppContext) WithHighlightFlags(cmd *cobra.Command) {
	cmd.Flags().BoolVarP(&app.NoHighlight, "no-highlight", "", false, "do not highlight output")
}

// WithPromptFlags sets up `cmd` for language based CLI flags.
func (app *AppContext) WithLanguageFlags(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&app.OutputLanguage, "language", "", "", "custom output language")
}

// WithPromptFlags sets up `cmd` for prompt based CLI flags.
func (app *AppContext) WithPromptFlags(cmd *cobra.Command) {
	app.WithEditorCLIFlags(cmd)
	app.WithHighlightFlags(cmd)
	app.WithSchemaFlags(cmd)
}

// WithSchemaFlags sets up `cmd` for (response) format based CLI flags.
func (app *AppContext) WithSchemaFlags(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&app.SchemaFile, "schema", "", "", "file with response format/schema")
	cmd.Flags().StringVarP(&app.SchemaName, "schema-name", "", "", "name of the response format/schema")
}
