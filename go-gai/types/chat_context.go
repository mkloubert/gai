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

	"github.com/goccy/go-yaml"
	"github.com/gosimple/slug"
)

// AppendSimplePseudoUserConversationOptions stores custom options for
// `AppendSimplePseudoUserConversation` of `ChatContext`
type AppendSimplePseudoUserConversationOptions struct {
	// Answer stores custom assistant answer.
	Answer *string
	// Model defines custom model to use.
	Model *string
	// Time defines custom time string to use.
	Time *string
}

// ChatContext handles chat context
type ChatContext struct {
	// App stores the underlying application context.
	App *AppContext
	// Conversations stores the repository with all conversations usually grouped by directory.
	Conversations  *ConversationRepository
	currentContext string
}

// UpdateConversationWith stores options for `UpdateConversationWith` method.
type UpdateConversationWithOptions struct {
	// NoSave is `true` if conversion file should not be updated.
	NoSave *bool
}

// AppendConversationItem adds a new item to the current converatio
// without updating the underyling conversation file.
func (ctx *ChatContext) AppendConversationItem(item *ConversationRepositoryConversationItem) *ConversationRepositoryConversationItem {
	conversationContext := ctx.ensureConversation()

	conversationContext.Conversation = append(conversationContext.Conversation, item)

	return item
}

// AppendSimplePseudoUserConversation adds an additional part of a pseudo conversation between a user and an assistant
// without updating the underyling conversation file.
func (ctx *ChatContext) AppendSimplePseudoUserConversation(um string, opts ...AppendSimplePseudoUserConversationOptions) []*ConversationRepositoryConversationItem {
	app := ctx.App

	answer := "OK"
	model := app.AI.ChatModel()
	time := app.GetISOTime()
	for _, o := range opts {
		if o.Answer != nil {
			answer = *o.Answer
		}
		if o.Model != nil {
			model = *o.Model
		}
		if o.Time != nil {
			time = *o.Time
		}
	}

	conversationContext := ctx.ensureConversation()

	newItems := make([]*ConversationRepositoryConversationItem, 0)

	// user message
	{
		userMessage := &ConversationRepositoryConversationItem{
			Contents: make(ConversationRepositoryConversationItemContents, 0),
			Model:    model,
			Role:     "user",
			Time:     time,
		}
		newUserTextItem := &ConversationRepositoryConversationItemContentItem{
			Content: um,
			Type:    "text",
		}
		userMessage.Contents = append(userMessage.Contents, newUserTextItem)

		newItems = append(newItems, ctx.AppendConversationItem(userMessage))
	}

	// simulate answer
	{
		assistantMessage := &ConversationRepositoryConversationItem{
			Contents: make(ConversationRepositoryConversationItemContents, 0),
			Model:    model,
			Role:     "assistant",
			Time:     time,
		}
		assistantMessage.Contents = append(assistantMessage.Contents, &ConversationRepositoryConversationItemContentItem{
			Content: answer,
			Type:    "text",
		})

		newItems = append(newItems, ctx.AppendConversationItem(assistantMessage))
	}

	conversationContext.Conversation = append(conversationContext.Conversation, newItems...)

	return newItems
}

func (ctx *ChatContext) ensureConversation() *ConversationRepositoryConversationContext {
	var app = ctx.App
	var cwd = app.WorkingDirectory
	var currentContext = ctx.currentContext

	repo := ctx.Conversations
	if repo == nil {
		repo = &ConversationRepository{}

		ctx.Conversations = repo
	}

	if repo.Conversations == nil {
		repo.Conversations = map[string]ConversationRepositoryConversationContextes{}
	}

	conversations, ok := repo.Conversations[cwd]
	if !ok || conversations == nil {
		conversations = ConversationRepositoryConversationContextes{}

		repo.Conversations[cwd] = conversations
	}

	context, ok := conversations[currentContext]
	if !ok || context == nil {
		context = &ConversationRepositoryConversationContext{}

		conversations[currentContext] = context
	}

	if context.Conversation == nil {
		context.Conversation = make(ConversationRepositoryConversation, 0)
	}

	return context
}

// GetConversation returns conversation for the current directory.
func (ctx *ChatContext) GetConversation() (ConversationRepositoryConversation, error) {
	conversationContext := ctx.ensureConversation()

	return conversationContext.Conversation, nil
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

	conversationContext := ctx.ensureConversation()
	conversationContext.Conversation = make(ConversationRepositoryConversation, 0)

	return nil
}

// SwitchContext switches the context.
func (ctx *ChatContext) SwitchContext(c string) string {
	newContextName := strings.TrimSpace(
		strings.ToLower(
			slug.MakeLang(c, "en"),
		),
	)

	ctx.currentContext = newContextName

	ctx.ensureConversation()

	return newContextName
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

	ctx.ensureConversation()

	data, err := yaml.Marshal(&ctx.Conversations)
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
func (ctx *ChatContext) UpdateConversationWith(conversation ConversationRepositoryConversation, opts ...UpdateConversationWithOptions) error {
	noSave := false
	for _, o := range opts {
		if o.NoSave != nil {
			noSave = *o.NoSave
		}
	}

	conversationContext := ctx.ensureConversation()
	conversationContext.Conversation = conversation

	if noSave {
		return nil
	}
	return ctx.UpdateConversation()
}
