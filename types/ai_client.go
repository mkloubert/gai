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

import "io"

// AIClient describes a client for an AI provider.
type AIClient interface {
	// AsSupportedAudioFormatString reads data as audio and tries to convert
	// it to a supported data format as data URI.
	AsSupportedAudioFormatString(b []byte) (string, error)
	// AsSupportedImageFormatString reads data as image and tries to convert
	// it to a supported data format as data URI.
	AsSupportedImageFormatString(b []byte) (string, error)
	// Chat starts or continues a chat conversation with message in `msg` based on `ctx` and returns the new conversation.
	Chat(ctx *ChatContext, msg string, opts ...AIClientChatOptions) (string, ConversationRepositoryConversation, error)
	// ChatModel returns the current chat model.
	ChatModel() string
	// Prompt does a single AI prompt with a specific `msg`.
	Prompt(msg string, opts ...AIClientPromptOptions) (AIClientPromptResponse, error)
	// Provider returns the name of the provider.
	Provider() string
	// SetChatModel sets the current chat model.
	SetChatModel(m string) error
}

// AIClientChatOptions stores additional options for `Chat` method.
type AIClientChatOptions struct {
	// Files stores list of one or more file to use for the submission.
	Files []io.Reader
}

// AIClientPromptOptions stores additional options for `Prompt` method.
type AIClientPromptOptions struct {
	// Files stores list of one or more file to use for the submission.
	Files []io.Reader
}

// AIClientPromptResponse stores information about a successful prompt response.
type AIClientPromptResponse struct {
	// Content stores the content.
	Content string
	// Model stores the model that has been used.
	Model string
}
