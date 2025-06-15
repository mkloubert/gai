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
)

// OllamaClient is an `AIClient` implementation for OpenAI.
type OpenAIClient struct {
	apiKey    string
	chatModel string
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
		if o.Files != nil {
			for _, f := range o.Files {
				if f != nil {
					data, err := io.ReadAll(f)
					if err != nil {
						return "", conversation, err
					}

					mimeType := http.DetectContentType(data)

					if strings.HasPrefix(mimeType, "image/") {
						dataURI, err := c.AsSupportedImageFormatString(data)
						if err != nil {
							return "", conversation, err
						}

						newUserImageItem := &ConversationRepositoryConversationItemContentItem{
							Content: dataURI,
							Type:    "image",
						}
						userMessage.Contents = append(userMessage.Contents, newUserImageItem)
					} else {
						return "", conversation, fmt.Errorf("mime type '%v' not supported", mimeType)
					}
				}
			}
		}
	}

	messages := []OpenAIChatMessage{}
	appendConversationItem := func(item *ConversationRepositoryConversationItem) error {
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
				}

				if newItem != nil {
					newMessage.Content = append(newMessage.Content, newItem)
				} else {
					return fmt.Errorf("content type '%v' not allowed", content.Type)
				}
			}

			messages = append(messages, *newMessage)
		}

		return nil
	}

	// add previous conversation
	for _, item := range conversation {
		err := appendConversationItem(item)
		if err != nil {
			return "", conversation, err
		}
	}
	// add user message
	appendConversationItem(userMessage)

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

	if resp.StatusCode != 200 {
		// load the response
		responseData, err := io.ReadAll(resp.Body)
		if err != nil {
			return "", conversation, err
		}

		return "", conversation, fmt.Errorf("unexpected response %v (%v)", resp.StatusCode, string(responseData))
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

// Provider returns the name of the provider.
func (c *OpenAIClient) Provider() string {
	return "openai"
}

// SetChatModel sets the current chat model.
func (c *OpenAIClient) SetChatModel(m string) error {
	c.chatModel = m
	return nil
}
