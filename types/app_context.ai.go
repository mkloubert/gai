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
	"errors"
	"fmt"
	"strings"
)

const initalOllamaChatModel = "llama3.1:8b"
const initalOpenAIChatModel = "gpt-4.1-mini"

// NewAIClient creates a new `AIClient` instance for a `provider`.
func (app *AppContext) NewAIClient(provider string) (AIClient, error) {
	provider = strings.TrimSpace(
		strings.ToLower(provider),
	)

	if provider == "ollama" {
		ollama := &OllamaClient{}
		ollama.chatModel = initalOllamaChatModel

		return ollama, nil
	}

	if provider == "openai" {
		apiKey := strings.TrimSpace(app.ApiKey)
		if apiKey == "" {
			apiKey = strings.TrimSpace(app.Getenv("OPENAI_API_KEY"))
		}
		if apiKey == "" {
			return nil, errors.New("no API defined")
		}

		openai := &OpenAIClient{}
		openai.chatModel = initalOpenAIChatModel
		openai.apiKey = apiKey

		return openai, nil
	}

	return nil, fmt.Errorf("'%v' is an unknown AI provider", provider)
}
