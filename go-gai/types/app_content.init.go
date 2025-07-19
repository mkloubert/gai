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
)

func (app *AppContext) initHomeDir() {
	// user's home directory
	homeDir, err := os.UserHomeDir()
	app.CheckIfError(err)

	if strings.TrimSpace(app.HomeDirectory) == "" {
		app.HomeDirectory = homeDir // default
	}

	// ensure absolute path
	if !filepath.IsAbs(app.HomeDirectory) {
		app.HomeDirectory = filepath.Join(homeDir, app.HomeDirectory)
	}
}

// InitAI initializes the default AI client.
func (app *AppContext) InitAI() {
	if strings.TrimSpace(app.Model) == "" {
		// first try command specific default model
		// GAI_DEFAULT_COMMAND_MODEL__*
		envSuffix := strings.Join(app.CommandPath, "_")
		envSuffix = strings.ToUpper(envSuffix)
		envName := fmt.Sprintf("GAI_DEFAULT_COMMAND_MODEL__%s", envSuffix)

		app.Model = strings.TrimSpace(
			app.GetEnv(envName),
		)
		if app.Model == "" {
			// now try env variable for common default
			app.Model = strings.TrimSpace(app.GetEnv("GAI_DEFAULT_CHAT_MODEL"))
		}
	}

	modelWithProvider := strings.TrimSpace(app.Model)
	if modelWithProvider == "" {
		app.CheckIfError(fmt.Errorf("no default chat model defined"))
	}

	sep := strings.Index(modelWithProvider, ":")
	if sep == -1 {
		app.CheckIfError(fmt.Errorf("no AI provider defined"))
	}

	provider := strings.TrimSpace(modelWithProvider[:sep])
	model := strings.TrimSpace(modelWithProvider[sep+1:])

	if provider == "" {
		app.CheckIfError(fmt.Errorf("no AI provider defined, use provider:model format"))
	}
	if model == "" {
		app.CheckIfError(fmt.Errorf("no chat model defined, use provider:model format"))
	}

	client, err := app.NewAIClient(provider)
	app.CheckIfError(err)

	app.Dbg(fmt.Sprintf("Using '%v' provider with '%v' model as default ...", provider, model))

	app.AI = client
}

func (app *AppContext) initWorkingDirectory() {
	// current working directory
	cwd, err := os.Getwd()
	app.CheckIfError(err)

	if strings.TrimSpace(app.WorkingDirectory) == "" {
		app.WorkingDirectory = cwd // default
	}

	// ensure absolute path
	if !filepath.IsAbs(app.WorkingDirectory) {
		app.WorkingDirectory = filepath.Join(cwd, app.WorkingDirectory)
	}
}

// Init initializes the application based on the current settings.
func (app *AppContext) Init() {
	app.initHomeDir()
	app.initWorkingDirectory()

	app.loadEnvFilesIfExist()

	app.loadRCFile()

	outputFile := app.GetOutputFile()
	if outputFile != "" {
		file, err := os.OpenFile(outputFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
		app.CheckIfError(err)

		app.Stdout = file
	}
}
