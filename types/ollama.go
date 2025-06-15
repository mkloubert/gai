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

// OllamaAIChatMessage is an item inside
// OllamaAIChat.Conversation array
type OllamaAIChatMessage struct {
	// Content stores the message content.
	Content string `json:"content,omitempty"`
	// Images stores URL of images to submit.
	Images []string `json:"images,omitempty"`
	// Role stores the role.
	Role string `json:"role,omitempty"`
}

// OllamaApiResponse is the data of a successful chat conversation response
type OllamaApiChatCompletionResponse struct {
	// Message stores the message.
	Message OllamaAIChatMessage `json:"message,omitempty"`
	// Model stores the model that has been used.
	Model string `json:"model,omitempty"`
}
