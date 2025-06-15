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
	"maps"
	"os"
	"path/filepath"
	"strings"

	"github.com/joho/godotenv"
)

// Getenv tries to return an environment variable
// and returns empty string if not found
func (app *AppContext) Getenv(key string) string {
	v, ok := app.EnvVars[key]
	if ok {
		return v
	}
	return ""
}

func (app *AppContext) loadEnvFilesIfExist() {
	envVars := map[string]string{}

	for _, env := range os.Environ() {
		parts := strings.SplitN(env, "=", 2)
		if len(parts) == 2 {
			key := parts[0]
			value := parts[1]

			envVars[key] = value
		}
	}

	loadFromFile := func(p string) {
		file, err := os.Open(p)
		app.CheckIfError(err)

		defer file.Close()

		vars, err := godotenv.Parse(file)
		app.CheckIfError(err)

		maps.Copy(envVars, vars)

		app.Dbg(fmt.Sprintf("Loaded .env file from %v", p))
	}

	if !app.SkipDefaultEnvFiles {
		// load default en files

		appDir, err := app.EnsureAppDir()
		app.CheckIfError(err)

		envPaths := []string{
			filepath.Join(app.HomeDirectory, ".env"),
			filepath.Join(appDir, ".env"),
			filepath.Join(app.WorkingDirectory, ".env"),
			filepath.Join(app.WorkingDirectory, ".env.local"),
		}

		for _, envPath := range envPaths {
			func() {
				if _, err := os.Stat(envPath); err == nil {
					loadFromFile(envPath)
				} else if !os.IsNotExist(err) {
					// could not check for .env
					app.CheckIfError(err)
				}
			}()
		}
	}

	if app.EnvFiles != nil {
		// custom and additional .env files

		for _, envFile := range app.EnvFiles {
			envPath := envFile
			if !filepath.IsAbs(envPath) {
				envPath = filepath.Join(app.WorkingDirectory, envPath)
			}

			loadFromFile(envPath)
		}
	}

	app.EnvVars = envVars
}
