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
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/mkloubert/gai/utils"
)

// OllamaClient is an `AIClient` implementation for OpenAI.
type OpenAIClient struct {
	apiKey    string
	app       *AppContext
	chatModel string
}

func (c *OpenAIClient) appendConversationItemTo(messages []OpenAIChatMessage, item *ConversationRepositoryConversationItem) ([]OpenAIChatMessage, error) {
	if item.Contents != nil {
		newMessage := &OpenAIChatMessage{
			Content: make(OpenAIChatMessageContent, 0),
			Role:    item.Role,
		}

		for _, content := range item.Contents {
			var newItem interface{}

			if content.Type == "text" {
				newItem = &OpenAIChatMessageContentTextItem{
					Text: content.Content,
					Type: "text",
				}
			} else if content.Type == "image" {
				newItem = &OpenAIChatMessageContentImageItem{
					ImageUrl: OpenAIChatMessageContentImageItemUrl{
						Url: content.Content,
					},
					Type: "image_url",
				}
			} else if content.Type == "audio" {
				parts := strings.SplitN(content.Content, ",", 2)
				if len(parts) != 2 {
					return messages, errors.New("invalid data URI")
				}

				meta := strings.TrimPrefix(parts[0], "data:")
				data := parts[1]

				// MIME-Typ extrahieren
				mimeParts := strings.SplitN(meta, ";", 2)
				if len(mimeParts) < 1 {
					return messages, fmt.Errorf("no MIME type found")
				}

				format := ""

				mime := strings.TrimSpace(
					strings.ToLower(mimeParts[0]),
				)
				if strings.HasSuffix(mime, "mp3") || strings.HasSuffix(mime, "mpeg") {
					format = "mp3"
				} else if strings.HasSuffix(mime, "wav") {
					format = "wav"
				}

				if format == "" {
					return messages, fmt.Errorf("unsupported audio format '%v'", mime)
				}

				newItem = &OpenAIChatMessageContentAudioItem{
					InputAudio: OpenAIChatMessageContentAudioItemInput{
						Data:   data,
						Format: format,
					},
					Type: "input_audio",
				}
			}

			if newItem != nil {
				newMessage.Content = append(newMessage.Content, newItem)
			} else {
				return messages, fmt.Errorf("content type '%v' not allowed", content.Type)
			}
		}

		messages = append(messages, *newMessage)
	}

	return messages, nil
}

func (c *OpenAIClient) appendFilesTo(item *ConversationRepositoryConversationItem, files []io.Reader) error {
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

				newUserImageItem := &ConversationRepositoryConversationItemContentItem{
					Content: dataURI,
					Type:    "image",
				}
				item.Contents = append(item.Contents, newUserImageItem)
			} else if strings.HasPrefix(mimeType, "audio/") {
				dataURI, err := c.AsSupportedAudioFormatString(data)
				if err != nil {
					return err
				}

				newUserImageItem := &ConversationRepositoryConversationItemContentItem{
					Content: dataURI,
					Type:    "audio",
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
func (c *OpenAIClient) AsSupportedAudioFormatString(b []byte) (string, error) {
	mimeType := http.DetectContentType(b)
	encoded := base64.StdEncoding.EncodeToString(b)
	dataURI := fmt.Sprintf("data:%s;base64,%s", mimeType, encoded)

	if strings.HasPrefix(mimeType, "audio/") {
		return dataURI, nil
	}
	return dataURI, fmt.Errorf("mime type '%v' is not a supported audio format", mimeType)
}

// AsSupportedImageFormatString reads data as image and tries to convert
// it to a supported data format as data URI.
func (c *OpenAIClient) AsSupportedImageFormatString(b []byte) (string, error) {
	mimeType := http.DetectContentType(b)
	encoded := base64.StdEncoding.EncodeToString(b)
	dataURI := fmt.Sprintf("data:%s;base64,%s", mimeType, encoded)

	if strings.HasPrefix(mimeType, "image/") {
		return dataURI, nil
	}
	return dataURI, fmt.Errorf("mime type '%v' is not a supported image format", mimeType)
}

// Chat starts or continues a chat conversation with message in `msg` based on `ctx` and returns the new conversation.
func (c *OpenAIClient) Chat(ctx *ChatContext, msg string, opts ...AIClientChatOptions) (string, ConversationRepositoryConversation, error) {
	conversation, err := ctx.GetConversation()
	if err != nil {
		return "", conversation, err
	}

	apiKey := strings.TrimSpace(c.apiKey)
	if apiKey == "" {
		return "", conversation, fmt.Errorf("no OpenAI api key defined")
	}

	model := strings.TrimSpace(strings.ToLower(c.chatModel))
	if model == "" {
		return "", conversation, fmt.Errorf("no chat ai model defined")
	}

	app := ctx.App

	maxTokens, err := app.GetMaxTokens()
	if err != nil {
		return "", conversation, err
	}

	temperature, err := app.GetTemperature()
	if err != nil {
		return "", conversation, err
	}

	baseUrl := app.GetBaseUrl()
	if baseUrl == "" {
		baseUrl = "https://api.openai.com" // use default
	}

	url := fmt.Sprintf("%v/v1/chat/completions", baseUrl)

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

	// add files
	for _, o := range opts {
		err := c.appendFilesTo(userMessage, o.Files)
		if err != nil {
			return "", conversation, err
		}
	}

	messages := []OpenAIChatMessage{}

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

	body := map[string]interface{}{
		"model":                 model,
		"messages":              messages,
		"stream":                false,
		"temperature":           temperature,
		"max_completion_tokens": maxTokens,
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
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiKey))
	// ... and finally send the JSON data
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", conversation, err
	}
	defer resp.Body.Close()

	err = utils.CheckForHttpResponseError(resp)
	if err != nil {
		return "", conversation, err
	}

	responseTime := app.GetISOTime()

	// load the response
	responseData, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", conversation, err
	}

	var chatResponse OpenAIChatCompletionResponseV1
	err = json.Unmarshal(responseData, &chatResponse)
	if err != nil {
		return "", conversation, err
	}

	answer := ""
	if len(chatResponse.Choices) > 0 {
		answer = chatResponse.Choices[0].Message.Content
	}

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
func (c *OpenAIClient) ChatModel() string {
	return c.chatModel
}

// Prompt does a single AI prompt with a specific `msg`.
func (c *OpenAIClient) Prompt(msg string, opts ...AIClientPromptOptions) (AIClientPromptResponse, error) {
	promptResponse := AIClientPromptResponse{
		Content: "",
		Model:   "",
	}

	apiKey := strings.TrimSpace(c.apiKey)
	if apiKey == "" {
		return promptResponse, fmt.Errorf("no OpenAI api key defined")
	}

	model := strings.TrimSpace(strings.ToLower(c.chatModel))
	if model == "" {
		return promptResponse, fmt.Errorf("no chat ai model defined")
	}

	promptResponse.Model = model

	app := c.app

	maxTokens, err := app.GetMaxTokens()
	if err != nil {
		return promptResponse, err
	}

	temperature, err := app.GetTemperature()
	if err != nil {
		return promptResponse, err
	}

	baseUrl := app.GetBaseUrl()
	if baseUrl == "" {
		baseUrl = "https://api.openai.com" // use default
	}

	url := fmt.Sprintf("%v/v1/chat/completions", baseUrl)

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

	// add files
	for _, o := range opts {
		err := c.appendFilesTo(userMessage, o.Files)
		if err != nil {
			return promptResponse, err
		}
	}

	messages := []OpenAIChatMessage{}

	// add user message
	m, err := c.appendConversationItemTo(messages, userMessage)
	if err != nil {
		return promptResponse, err
	}

	messages = m

	body := map[string]interface{}{
		"model":                 model,
		"messages":              messages,
		"stream":                false,
		"temperature":           temperature,
		"max_completion_tokens": maxTokens,
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
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiKey))
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

	var chatResponse OpenAIChatCompletionResponseV1
	err = json.Unmarshal(responseData, &chatResponse)
	if err != nil {
		return promptResponse, err
	}

	answer := ""
	if len(chatResponse.Choices) > 0 {
		answer = chatResponse.Choices[0].Message.Content
	}

	promptResponse.Content = answer
	promptResponse.Model = chatResponse.Model

	return promptResponse, nil
}

// Provider returns the name of the provider.
func (c *OpenAIClient) Provider() string {
	return "openai"
}

// SetChatModel sets the current chat model.
func (c *OpenAIClient) SetChatModel(m string) error {
	c.chatModel = m
	return nil
}
