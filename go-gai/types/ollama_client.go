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
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/mkloubert/gai/utils"
)

// OllamaClient is an `AIClient` implementation for Ollama.
type OllamaClient struct {
	app       *AppContext
	chatModel string
}

func (c *OllamaClient) appendConversationItemTo(messages []OllamaAIChatMessage, item *ConversationRepositoryConversationItem) ([]OllamaAIChatMessage, error) {
	if item.Contents != nil {
		newMessage := &OllamaAIChatMessage{
			Content: "",
			Images:  make([]string, 0),
			Role:    item.Role,
		}

		for _, content := range item.Contents {
			if content.Type == "text" {
				newMessage.Content = content.Content
			} else if content.Type == "image" {
				newMessage.Images = append(newMessage.Images, content.Content)
			} else {
				return messages, fmt.Errorf("content type '%v' not allowed", content.Type)
			}
		}

		messages = append(messages, *newMessage)
	}

	return messages, nil
}

func (c *OllamaClient) appendFilesTo(item *ConversationRepositoryConversationItem, files []io.Reader) error {
	for _, f := range files {
		if f != nil {
			data, err := io.ReadAll(f)
			if err != nil {
				return err
			}

			mimeType := http.DetectContentType(data)

			if strings.HasPrefix(mimeType, "image/") {
				dataURI, err := c.AsSupportedImageFormatString(data)
				if err != nil {
					return err
				}

				comma := strings.Index(dataURI, ",")

				newUserImageItem := &ConversationRepositoryConversationItemContentItem{
					Content: dataURI[comma+1:],
					Type:    "image",
				}
				item.Contents = append(item.Contents, newUserImageItem)
			} else {
				return fmt.Errorf("mime type '%v' not supported", mimeType)
			}
		}
	}

	return nil
}

// AsSupportedAudioFormatString reads data as audio and tries to convert
// it to a supported data format as data URI.
func (c *OllamaClient) AsSupportedAudioFormatString(b []byte) (string, error) {
	mimeType := http.DetectContentType(b)

	return "", fmt.Errorf("mime type '%v' is not a supported audio format", mimeType)
}

// AsSupportedImageFormatString reads data as image and tries to convert
// it to a supported data format as data URI.
func (c *OllamaClient) AsSupportedImageFormatString(b []byte) (string, error) {
	mimeType := http.DetectContentType(b)
	encoded := base64.StdEncoding.EncodeToString(b)
	dataURI := fmt.Sprintf("data:%s;base64,%s", mimeType, encoded)

	if strings.HasPrefix(mimeType, "image/") {
		return dataURI, nil
	}
	return dataURI, fmt.Errorf("mime type '%v' is not a supported image format", mimeType)
}

// Chat starts or continues a chat conversation with message in `msg` based on `ctx` and returns the new conversation.
func (c *OllamaClient) Chat(ctx *ChatContext, msg string, opts ...AIClientChatOptions) (string, ConversationRepositoryConversation, error) {
	conversation, err := ctx.GetConversation()
	if err != nil {
		return "", conversation, err
	}

	model := strings.TrimSpace(strings.ToLower(c.chatModel))
	if model == "" {
		return "", conversation, fmt.Errorf("no chat ai model defined")
	}

	conversation = c.setupSystemPromptIfNeeded(conversation, "", model)

	app := ctx.App

	temperature, err := app.GetTemperature()
	if err != nil {
		return "", conversation, err
	}

	baseUrl := app.GetBaseUrl()
	if baseUrl == "" {
		baseUrl = "http://localhost:11434" // use default
	}

	var schema *map[string]any
	for _, o := range opts {
		if o.ResponseSchema != nil {
			schema = o.ResponseSchema
		}
	}

	url := fmt.Sprintf("%v/api/chat", baseUrl)

	userMessage := &ConversationRepositoryConversationItem{
		Contents: make(ConversationRepositoryConversationItemContents, 0),
		Model:    model,
		Role:     "user",
	}
	newUserTextItem := &ConversationRepositoryConversationItemContentItem{
		Content: msg,
		Type:    "text",
	}
	userMessage.Contents = append(userMessage.Contents, newUserTextItem)

	// add response format
	responseFormat, err := c.writeResponseFormatTo(userMessage, schema)
	if err != nil {
		return "", conversation, err
	}

	// add files
	for _, o := range opts {
		if o.Files == nil {
			continue
		}

		err := c.appendFilesTo(userMessage, *o.Files)
		if err != nil {
			return "", conversation, err
		}
	}

	messages := []OllamaAIChatMessage{}

	// add previous conversation
	for _, item := range conversation {
		m, err := c.appendConversationItemTo(messages, item)
		if err != nil {
			return "", conversation, err
		}

		messages = m
	}

	// add user message
	m, err := c.appendConversationItemTo(messages, userMessage)
	if err != nil {
		return "", conversation, err
	}

	messages = m

	body := map[string]any{
		"model":    c.chatModel,
		"messages": messages,
		"stream":   false,
		"options": map[string]any{
			"temperature": temperature,
		},
		"format": responseFormat,
	}

	jsonData, err := json.Marshal(&body)
	if err != nil {
		return "", conversation, err
	}

	userMessage.Time = app.GetISOTime()

	req, err := http.NewRequest("POST", url, bytes.NewBuffer([]byte(jsonData)))
	if err != nil {
		return "", conversation, err
	}

	// setup ...
	req.Header.Set("Content-Type", "application/json")
	// ... and finally send the JSON data
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", conversation, err
	}
	defer resp.Body.Close()

	responseTime := app.GetISOTime()

	err = utils.CheckForHttpResponseError(resp)
	if err != nil {
		return "", conversation, err
	}

	// load the response
	responseData, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", conversation, err
	}

	var chatResponse OllamaApiChatCompletionResponse
	err = json.Unmarshal(responseData, &chatResponse)
	if err != nil {
		return "", conversation, err
	}

	answer := chatResponse.Message.Content

	// update conversation
	{
		conversation = append(conversation, userMessage)

		// take assistant message
		assistantMessage := &ConversationRepositoryConversationItem{
			Contents: make(ConversationRepositoryConversationItemContents, 0),
			Model:    chatResponse.Model,
			Role:     "assistant",
			Time:     responseTime,
		}
		assistantMessage.Contents = append(assistantMessage.Contents, &ConversationRepositoryConversationItemContentItem{
			Content: answer,
			Type:    "text",
		})
		conversation = append(conversation, assistantMessage)
	}

	ctx.UpdateConversationWith(conversation)

	return answer, conversation, nil
}

// ChatModel returns the current chat model.
func (c *OllamaClient) ChatModel() string {
	return c.chatModel
}

// Prompt does a single AI prompt with a specific `msg`.
func (c *OllamaClient) Prompt(msg string, opts ...AIClientPromptOptions) (AIClientPromptResponse, error) {
	promptResponse := AIClientPromptResponse{
		Content: "",
		Model:   "",
	}

	model := strings.TrimSpace(strings.ToLower(c.chatModel))
	if model == "" {
		return promptResponse, fmt.Errorf("no chat ai model defined")
	}

	tempConversation := make(ConversationRepositoryConversation, 0)
	tempConversation = c.setupSystemPromptIfNeeded(tempConversation, "", model)

	promptResponse.Model = model

	app := c.app

	temperature, err := app.GetTemperature()
	if err != nil {
		return promptResponse, err
	}

	baseUrl := app.GetBaseUrl()
	if baseUrl == "" {
		baseUrl = "http://localhost:11434" // use default
	}

	var schema *map[string]any
	for _, o := range opts {
		if o.ResponseSchema != nil {
			schema = o.ResponseSchema
		}
	}

	url := fmt.Sprintf("%v/api/generate", baseUrl)

	userMessage := &ConversationRepositoryConversationItem{
		Contents: make(ConversationRepositoryConversationItemContents, 0),
		Model:    model,
		Role:     "user",
	}
	newUserTextItem := &ConversationRepositoryConversationItemContentItem{
		Content: msg,
		Type:    "text",
	}
	userMessage.Contents = append(userMessage.Contents, newUserTextItem)

	// add response format
	responseFormat, err := c.writeResponseFormatTo(userMessage, schema)
	if err != nil {
		return promptResponse, err
	}

	// add files
	for _, o := range opts {
		if o.Files == nil {
			continue
		}

		err := c.appendFilesTo(userMessage, *o.Files)
		if err != nil {
			return promptResponse, err
		}
	}

	messages := []OllamaAIChatMessage{}
	for _, item := range tempConversation {
		m, err := c.appendConversationItemTo(messages, item)
		if err != nil {
			return promptResponse, err
		}

		messages = m
	}

	m, err := c.appendConversationItemTo(messages, userMessage)
	if err != nil {
		return promptResponse, err
	}

	messages = m

	images := make([]string, 0)
	for i, c := range userMessage.Contents {
		if i < 1 {
			continue
		}

		if c.Type == "image" {
			images = append(images, c.Content)
		} else {
			return promptResponse, fmt.Errorf("content type '%v' not supported", c.Type)
		}
	}

	body := map[string]any{
		"model":       model,
		"prompt":      userMessage.Contents[0].Content,
		"stream":      false,
		"temperature": temperature,
		"images":      images,
		"format":      responseFormat,
	}

	jsonData, err := json.Marshal(&body)
	if err != nil {
		return promptResponse, err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer([]byte(jsonData)))
	if err != nil {
		return promptResponse, err
	}

	// setup ...
	req.Header.Set("Content-Type", "application/json")
	// ... and finally send the JSON data
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return promptResponse, err
	}
	defer resp.Body.Close()

	err = utils.CheckForHttpResponseError(resp)
	if err != nil {
		return promptResponse, err
	}

	// load the response
	responseData, err := io.ReadAll(resp.Body)
	if err != nil {
		return promptResponse, err
	}

	var completionResponse OllamaApiCompletionResponse
	err = json.Unmarshal(responseData, &completionResponse)
	if err != nil {
		return promptResponse, err
	}

	answer := completionResponse.Response

	promptResponse.Content = answer
	promptResponse.Model = completionResponse.Model

	return promptResponse, nil
}

// Provider returns the name of the provider.
func (c *OllamaClient) Provider() string {
	return "ollama"
}

// SetChatModel sets the current chat model.
func (c *OllamaClient) SetChatModel(m string) error {
	c.chatModel = m
	return nil
}

func (c *OllamaClient) setupSystemPromptIfNeeded(conversation ConversationRepositoryConversation, defaultPrompt string, model string) ConversationRepositoryConversation {
	if len(conversation) == 0 {
		// only if no conversation yet ...

		app := c.app

		systemPrompt := strings.TrimSpace(
			app.GetSystemPrompt(defaultPrompt),
		)
		if systemPrompt != "" {
			// ... system prompt is defined

			systemMessage := &ConversationRepositoryConversationItem{
				Contents: make(ConversationRepositoryConversationItemContents, 0),
				Model:    model,
				Role:     app.GetSystemRole(),
				Time:     app.GetISOTime(),
			}
			newTextItem := &ConversationRepositoryConversationItemContentItem{
				Content: systemPrompt,
				Type:    "text",
			}
			systemMessage.Contents = append(systemMessage.Contents, newTextItem)

			conversation = append(conversation, systemMessage)
		}
	}

	return conversation
}

func (c *OllamaClient) toResponseFormat(schema *map[string]any) *map[string]any {
	return schema
}

func (c *OllamaClient) writeResponseFormatTo(item *ConversationRepositoryConversationItem, schema *map[string]any) (*map[string]any, error) {
	responseFormat := c.toResponseFormat(schema)
	if responseFormat != nil {
		jsonData, err := json.Marshal(responseFormat)
		if err != nil {
			return responseFormat, err
		}

		item.ResponseFormat = string(jsonData)
	}

	return responseFormat, nil
}
