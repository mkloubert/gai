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

	"github.com/goccy/go-yaml"
)

// ChatContext handles chat context
type ChatContext struct {
	// App stores the underlying application context.
	App *AppContext
	// Conversations stores the repository with all conversations usually grouped by directory.
	Conversations *ConversationRepository
}

func (ctx *ChatContext) ensureConversations() *ConversationRepository {
	repo := ctx.Conversations
	if repo == nil {
		repo = &ConversationRepository{}

		ctx.Conversations = repo
	}

	return repo
}

// GetConversation returns conversation for the current directory.
func (ctx *ChatContext) GetConversation() (ConversationRepositoryConversation, error) {
	app := ctx.App

	appConversations := make(ConversationRepositoryConversation, 0)

	if ctx.Conversations != nil {
		items, ok := ctx.Conversations.Conversations[app.WorkingDirectory]
		if ok {
			appConversations = items
		}
	}

	return appConversations, nil
}

func (ctx *ChatContext) getConversaionsFilePath() (string, error) {
	app := ctx.App

	appDir, err := app.EnsureAppDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(appDir, ".conversations.yaml"), nil
}

// ReloadAllConversations reloads the conversation file with all conversations
// and writes it to `Conversations`.
func (ctx *ChatContext) ReloadAllConversations() error {
	conversationFile, err := ctx.getConversaionsFilePath()
	if err != nil {
		return err
	}

	app := ctx.App

	var repo ConversationRepository

	if _, err := os.Stat(conversationFile); err == nil {
		app.Dbg(fmt.Sprintf("Loading conversations from '%v' ...", conversationFile))

		file, err := os.Open(conversationFile)
		if err != nil {
			return err
		}

		defer file.Close()

		dec := yaml.NewDecoder(file)

		err = dec.Decode(&repo)
		if err != nil {
			return err
		}
	} else if !os.IsNotExist(err) {
		// could not check for conversationFile
		return err
	} else {
		app.Dbg("Conversation file not found")
	}

	ctx.Conversations = &repo
	return nil
}

// ResetConversation resets conversation of the current directory.
func (ctx *ChatContext) ResetConversation() error {
	app := ctx.App

	app.Dbg(fmt.Sprintf("Resetting conversation of %v ...", app.WorkingDirectory))

	if ctx.Conversations == nil {
		ctx.Conversations = &ConversationRepository{}
	}
	if ctx.Conversations.Conversations == nil {
		ctx.Conversations.Conversations = map[string]ConversationRepositoryConversation{}
	}

	ctx.Conversations.Conversations[app.WorkingDirectory] = make(ConversationRepositoryConversation, 0)

	return nil
}

// UpdateConversation updates the conversation file with all conversations.
func (ctx *ChatContext) UpdateConversation() error {
	conversationFile, err := ctx.getConversaionsFilePath()
	if err != nil {
		return err
	}

	app := ctx.App

	app.Dbg(fmt.Sprintf("Will write conversations to '%v' ...", conversationFile))
	app.Dbg("Creating YAML data ...")

	repo := ctx.ensureConversations()

	data, err := yaml.Marshal(&repo)
	if err != nil {
		return err
	}

	app.Dbg(fmt.Sprintf("Writing conversations to '%v' ...", conversationFile))

	err = os.WriteFile(conversationFile, data, 0644)
	if err != nil {
		return err
	}
	return nil
}

// UpdateConversationWith updates the conversation file with all conversations.
func (ctx *ChatContext) UpdateConversationWith(conversation ConversationRepositoryConversation) error {
	app := ctx.App

	repo := ctx.ensureConversations()
	if repo.Conversations == nil {
		repo.Conversations = map[string]ConversationRepositoryConversation{}
	}

	if conversation != nil {
		repo.Conversations[app.WorkingDirectory] = conversation
	} else {
		repo.Conversations[app.WorkingDirectory] = make(ConversationRepositoryConversation, 0)
	}

	return ctx.UpdateConversation()
}
