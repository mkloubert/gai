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

import "fmt"

// AIModel manages an AI model for an `AIClient`.
type AIModel struct {
	client    AIClient
	modelType string
	name      string
}

// Client returns the AI client instance.
func (m *AIModel) Client() AIClient {
	return m.client
}

// Name returns the name of the model without provider prefix.
func (m *AIModel) Name() string {
	return m.name
}

// ModelType returns the type of the model.
func (m *AIModel) ModelType() string {
	return m.modelType
}

// NewAIModel creates a new instance of an `AIModel` for an AI client.
func NewAIModel(client AIClient, name string, modelType string) *AIModel {
	return &AIModel{
		client: client,
		name:   name,
	}
}

// String returns the full name (with provider prefix).
func (m *AIModel) String() string {
	return fmt.Sprintf("%s:%s", m.client.Provider(), m.name)
}
