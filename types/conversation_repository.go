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

// ConversationRepository represents a file of an AI conversation.
type ConversationRepository struct {
	// Conversations stores the conversation for each directory.
	// The key is the ID, usually the current directory, of the conversation.
	Conversations map[string]ConversationRepositoryConversationContextes `yaml:"conversations"`
}

// ConversationRepositoryConversation is a list of `ConversationRepositoryConversationItem`.
type ConversationRepositoryConversation = []*ConversationRepositoryConversationItem

// ConversationRepositoryConversationItem is an item in a `ConversationRepositoryConversation`.
type ConversationRepositoryConversationItem struct {
	// Content stores the usually Markdown content of the item.
	Contents ConversationRepositoryConversationItemContents `yaml:"contents"`
	// Model stores the model that is/has been used.
	Model string `yaml:"model"`
	// Role stores the role like `assistant`, `system` or `user`
	Role string `yaml:"role"`
	// Time stores timestamp in ISO 8601 format.
	Time string `yaml:"time"`
}

// ConversationRepositoryConversationItemContents stores list of `ConversationRepositoryConversationItemContentItem`s.
type ConversationRepositoryConversationItemContents = []*ConversationRepositoryConversationItemContentItem

// ConversationRepositoryConversationItemContentItem is an item inside a `ConversationRepositoryConversationItemContentItem`.
type ConversationRepositoryConversationItemContentItem struct {
	// Content stores the string serialized content.
	Content string `yaml:"content"`
	// Type stores the type like `text` or `image`.
	Type string `yaml:"type"`
}

// ConversationRepositoryConversationContext stores a conversation in a specific context.
type ConversationRepositoryConversationContext struct {
	// Conversation stores the underlying conversation.
	Conversation ConversationRepositoryConversation `yaml:"conversation"`
}

// ConversationRepositoryConversationContextes stores contextes grouped by their name/ID.
type ConversationRepositoryConversationContextes map[string]*ConversationRepositoryConversationContext
