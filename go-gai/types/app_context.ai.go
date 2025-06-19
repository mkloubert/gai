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
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"golang.org/x/term"
)

const initalOllamaChatModel = "ollama:llama3.1:8b"
const initalOpenAIChatModel = "openai:gpt-4.1-mini"

// GetBaseUrl returns the base URL for API operations, if defined.
func (app *AppContext) GetBaseUrl() string {
	baseUrl := strings.TrimSpace(app.BaseUrl)
	if baseUrl == "" {
		baseUrl = strings.TrimSpace(app.GetEnv("GAI_BASE_URL"))
	}

	return baseUrl
}

// GetMaxTokens returns the maximum number of GPT tokens to return / use.
func (app *AppContext) GetMaxTokens() (*int64, error) {
	maxTokens := app.MaxTokens

	if maxTokens <= 0 {
		GAI_MAX_TOKENS := strings.TrimSpace(app.GetEnv("GAI_MAX_TOKENS"))
		if GAI_MAX_TOKENS != "" {
			num, err := strconv.ParseInt(GAI_MAX_TOKENS, 10, 64)
			if err != nil {
				return nil, err
			}

			maxTokens = num
		}
	}

	if maxTokens > 0 {
		return &maxTokens, nil
	}
	return nil, nil // let AI decide to use what default
}

// GetResponseSchema loads the data for response format with schema and name.
func (app *AppContext) GetResponseSchema() (*map[string]any, string, error) {
	var schema *map[string]any
	schemaName := ""

	schemaFile := strings.TrimSpace(app.SchemaFile)
	if schemaFile != "" {
		if !filepath.IsAbs(schemaFile) {
			schemaFile = filepath.Join(app.WorkingDirectory, schemaFile)
		}

		schemaName = strings.TrimSpace(app.SchemaName)
		if schemaName == "" {
			schemaName = "GaiResponseSchema"
		}

		file, err := os.Open(schemaFile)
		if err != nil {
			return schema, schemaName, err
		}
		defer file.Close()

		data, err := io.ReadAll(file)
		if err != nil {
			return schema, schemaName, err
		}

		var temp map[string]any
		err = json.Unmarshal(data, &temp)
		if err != nil {
			return schema, schemaName, err
		}

		schema = &temp
	}

	return schema, schemaName, nil
}

// GetSystemPrompt returns the system prompt value for AI operations.
func (app *AppContext) GetSystemPrompt(defaultPrompt string) string {
	systemPrompt := strings.TrimSpace(app.SystemPrompt) // first try flag
	if systemPrompt == "" {
		systemPrompt = strings.TrimSpace(app.GetEnv("GAI_SYSTEM_PROMPT")) // now try env variable
	}

	if systemPrompt == "" {
		return defaultPrompt
	}
	return systemPrompt
}

// GetSystemRole returns the name/ID of the system role for AI operations.
func (app *AppContext) GetSystemRole() string {
	systemRole := strings.TrimSpace(app.SystemRole) // first try flag
	if systemRole == "" {
		systemRole = strings.TrimSpace(app.GetEnv("GAI_SYSTEM_ROLE")) // now try env variable
	}

	if systemRole == "" {
		systemRole = "system"
	}

	return systemRole
}

// GetTemperature returns the temperature value for AI operations.
func (app *AppContext) GetTemperature() (float64, error) {
	if app.Temperature >= 0 {
		return app.Temperature, nil
	}

	GAI_TEMPERATURE := strings.TrimSpace(
		app.GetEnv("GAI_TEMPERATURE"),
	)
	if GAI_TEMPERATURE != "" {
		f64, err := strconv.ParseFloat(GAI_TEMPERATURE, 64)
		if err != nil {
			return -2, err
		}

		return f64, nil
	}

	return 0.3, nil
}

// NewAIClient creates a new `AIClient` instance for a `provider`.
func (app *AppContext) NewAIClient(provider string) (AIClient, error) {
	provider = strings.TrimSpace(
		strings.ToLower(provider),
	)

	getModelNameOnly := func(m string) (string, error) {
		prefix := fmt.Sprintf("%v:", provider)
		if strings.HasPrefix(m, prefix) {
			return strings.TrimSpace(m[len(prefix):]), nil
		}

		return m, errors.New("could not get model format")
	}

	if provider == "ollama" {
		m := strings.TrimSpace(app.Model)
		if m == "" {
			m = initalOllamaChatModel
		}

		chatModel, err := getModelNameOnly(m)
		if err != nil {
			return nil, err
		}

		ollama := &OllamaClient{}
		ollama.app = app
		ollama.chatModel = chatModel

		return ollama, nil
	}

	if provider == "openai" {
		m := strings.TrimSpace(app.Model)
		if m == "" {
			m = initalOpenAIChatModel
		}

		chatModel, err := getModelNameOnly(m)
		if err != nil {
			return nil, err
		}

		apiKey := strings.TrimSpace(app.ApiKey)
		if apiKey == "" {
			// now try env variable
			apiKey = strings.TrimSpace(app.GetEnv("OPENAI_API_KEY"))
		}
		if apiKey == "" {
			return nil, errors.New("no API defined")
		}

		openai := &OpenAIClient{}
		openai.apiKey = apiKey
		openai.app = app
		openai.chatModel = chatModel

		return openai, nil
	}

	return nil, fmt.Errorf("'%v' is an unknown AI provider", provider)
}

// OutputAIAnswer outputs an AI answer to STDOUT.
func (app *AppContext) OutputAIAnswer(answer string) {
	stdout := app.Stdout

	if !app.NoHighlight && term.IsTerminal(int(stdout.Fd())) {
		chroma := app.GetChromaSettings()
		chroma.HighlightMarkdown(answer)

		app.Writeln()
	} else {
		app.WriteString(answer)
	}
}
