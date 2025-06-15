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

// OpenAIChatCompletionResponseV1 stores data of a successful
// OpenAI chat completion response API response (version 1).
type OpenAIChatCompletionResponseV1 struct {
	// Choices contains list of choices.
	Choices []OpenAIChatCompletionResponseV1Choice `json:"choices"`
	// Model stores the used model.
	Model string `json:"model"`
	// Usage stores the used resources.
	Usage OpenAIChatCompletionResponseV1Usage `json:"usage"`
}

// OpenAIChatCompletionResponseV1Choice is an item inside `choices` property
// of an `OpenAIChatCompletionResponseV1` object.
type OpenAIChatCompletionResponseV1Choice struct {
	// Index stores the zero-based index.
	Index int32 `json:"index"`
	// Message stores the message information.
	Message OpenAIChatCompletionResponseV1ChoiceMessage `json:"message"`
}

// OpenAIChatCompletionResponseV1ChoiceMessage contains data for `message` property
// of an `OpenAIChatCompletionResponseV1ChoiceMessage` object.
type OpenAIChatCompletionResponseV1ChoiceMessage struct {
	// Content stores the message content.
	Content string `json:"content"`
	// Stores the role like 'system' , 'user' or 'assistant'
	Role string `json:"role"`
}

// OpenAIChatCompletionResponseV1Usage contains data for `usage` property
// of an `OpenAIChatCompletionResponseV1` object
type OpenAIChatCompletionResponseV1Usage struct {
	// CompletionTokens stores number of completion tokens.
	CompletionTokens int32 `json:"completion_tokens"`
	// PromptTokens stores number of prompt tokens.
	PromptTokens int32 `json:"prompt_tokens"`
	// TotalTokens stores number of total used tokens.
	TotalTokens int32 `json:"total_tokens"`
}

// OpenAIChatMessage stores data of an OpenAI client chat message.
type OpenAIChatMessage struct {
	// Content stores the message content.
	Content OpenAIChatMessageContent `json:"content,omitempty"`
	// Role stores the role.
	Role string `json:"role,omitempty"`
}

// OpenAIChatMessageContent stores list of `OpenAIChatMessageContentItem`s.
type OpenAIChatMessageContent = []OpenAIChatMessageContentItem

// OpenAIChatMessageContentItem is an item inside an `OpenAIChatMessageContent`.
type OpenAIChatMessageContentItem = interface{}

// OpenAIChatMessageContentTextItem represents an `OpenAIChatMessageContentItem` of type `text`.
type OpenAIChatMessageContentTextItem struct {
	// Text stores the message content.
	Text string `json:"text,omitempty"`
	// Type stores the value `text`.
	Type string `json:"type,omitempty"`
}
