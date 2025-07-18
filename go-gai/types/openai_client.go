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
	"mime"
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

type openaiGetModelListResponse struct {
	Data []openaiGetModelListItem `json:"data"`
}

type openaiGetModelListItem struct {
	Id      string `json:"id"`
	Object  string `json:"object"`
	OwnedBy string `json:"owned_by"`
}

func (c *OpenAIClient) appendConversationItemTo(messages []OpenAIChatMessage, item *ConversationRepositoryConversationItem) ([]OpenAIChatMessage, error) {
	if item.Contents != nil {
		newMessage := &OpenAIChatMessage{
			Content: make(OpenAIChatMessageContent, 0),
			Role:    item.Role,
		}

		for i, content := range item.Contents {
			var newItem any

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
				_, mimeType, err := utils.GetPartsOfDataURI(content.Content)
				if err != nil {
					return messages, err
				}

				format := ""

				if strings.HasSuffix(mimeType, "mp3") || strings.HasSuffix(mimeType, "mpeg") {
					format = "mp3"
				} else if strings.HasSuffix(mimeType, "wav") {
					format = "wav"
				}

				if format == "" {
					return messages, fmt.Errorf("unsupported audio format '%v'", mimeType)
				}

				newItem = &OpenAIChatMessageContentAudioItem{
					InputAudio: OpenAIChatMessageContentItemInputAudio{
						Data:   content.Content,
						Format: format,
					},
					Type: "input_audio",
				}
			} else {
				// handle as file attachment

				_, mimeType, err := utils.GetPartsOfDataURI(content.Content)
				if err != nil {
					return messages, err
				}

				fileExt, err := mime.ExtensionsByType(mimeType)
				if err != nil {
					return messages, err
				}

				newItem = &OpenAIChatMessageContentFileItem{
					File: OpenAIChatMessageContentItemFile{
						FileData: content.Content,
						Filename: fmt.Sprintf("file_%d%s", i+1, fileExt),
					},
					Type: "file",
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

			mimeType := utils.DetectMime(data)

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
				mimeType := utils.DetectMime(data)
				encoded := base64.StdEncoding.EncodeToString(data)
				dataURI := fmt.Sprintf("data:%s;base64,%s", mimeType, encoded)

				newUserImageItem := &ConversationRepositoryConversationItemContentItem{
					Content: dataURI,
					Type:    "attachment",
				}
				item.Contents = append(item.Contents, newUserImageItem)
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
	mimeType := utils.DetectMime(b)
	encoded := base64.StdEncoding.EncodeToString(b)
	dataURI := fmt.Sprintf("data:%s;base64,%s", mimeType, encoded)

	if strings.HasPrefix(mimeType, "image/") {
		if strings.HasSuffix(mimeType, "/jpeg") || strings.HasSuffix(mimeType, "/jpg") || strings.HasSuffix(mimeType, "/png") {
			return dataURI, nil
		}

		pngData, err := utils.EnsurePNG(b)
		if err == nil {
			encoded = base64.StdEncoding.EncodeToString(pngData)
			dataURI = fmt.Sprintf("data:%s;base64,%s", "image/png", encoded)
		}

		return dataURI, err
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

	noSave := false
	systemPrompt := ""
	for _, o := range opts {
		if o.NoSave != nil {
			noSave = *o.NoSave
		}
		if o.SystemPrompt != nil {
			systemPrompt = *o.SystemPrompt
		}
	}

	conversation = c.setupSystemPromptIfNeeded(conversation, systemPrompt, model)

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

	var schema *map[string]any
	schemaName := ""
	for _, o := range opts {
		if o.ResponseSchema != nil {
			schema = o.ResponseSchema
		}
		if o.ResponseSchemaName != nil {
			schemaName = *o.ResponseSchemaName
		}
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

	// add response format
	responseFormat, err := c.writeResponseFormatTo(userMessage, schema, schemaName)
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

	body := map[string]any{
		"model":                 model,
		"messages":              messages,
		"stream":                false,
		"temperature":           temperature,
		"max_completion_tokens": maxTokens,
		"response_format":       responseFormat,
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

	if !noSave {
		err := ctx.UpdateConversationWith(conversation)
		if err != nil {
			return answer, conversation, err
		}
	}

	return answer, conversation, nil
}

// ChatModel returns the current chat model.
func (c *OpenAIClient) ChatModel() string {
	return c.chatModel
}

// Returns the list of supported OpenAI models.
func (c *OpenAIClient) GetModels() ([]AIModel, error) {
	models := make([]AIModel, 0)

	apiKey := strings.TrimSpace(c.apiKey)
	if apiKey == "" {
		return models, fmt.Errorf("no OpenAI api key defined")
	}

	app := c.app

	baseUrl := app.GetBaseUrl()
	if baseUrl == "" {
		baseUrl = "https://api.openai.com" // use default
	}

	url := fmt.Sprintf("%s/v1/models", baseUrl)

	req, err := http.NewRequest("GET", url, bytes.NewBuffer([]byte{}))
	if err != nil {
		return models, err
	}

	// setup
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", apiKey))

	// ... and finally send the JSON data
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return models, err
	}
	defer resp.Body.Close()

	err = utils.CheckForHttpResponseError(resp)
	if err != nil {
		return models, err
	}

	// load the response
	responseData, err := io.ReadAll(resp.Body)
	if err != nil {
		return models, err
	}

	var listResponse openaiGetModelListResponse
	err = json.Unmarshal(responseData, &listResponse)
	if err != nil {
		return models, err
	}

	for _, item := range listResponse.Data {
		if item.OwnedBy != "openai" && item.OwnedBy != "system" {
			continue
		}

		models = append(models, AIModel{
			client:    c,
			modelType: "",
			name:      item.Id,
		})
	}

	return models, nil
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

	systemPrompt := ""
	for _, o := range opts {
		if o.SystemPrompt != nil {
			systemPrompt = *o.SystemPrompt
		}
	}

	tempConversation := make(ConversationRepositoryConversation, 0)
	tempConversation = c.setupSystemPromptIfNeeded(tempConversation, systemPrompt, model)

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

	var schema *map[string]any
	schemaName := ""
	for _, o := range opts {
		if o.ResponseSchema != nil {
			schema = o.ResponseSchema
		}
		if o.ResponseSchemaName != nil {
			schemaName = *o.ResponseSchemaName
		}
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

	// add response format
	responseFormat, err := c.writeResponseFormatTo(userMessage, schema, schemaName)
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

	messages := []OpenAIChatMessage{}
	for _, item := range tempConversation {
		m, err := c.appendConversationItemTo(messages, item)
		if err != nil {
			return promptResponse, err
		}

		messages = m
	}

	// add user message
	m, err := c.appendConversationItemTo(messages, userMessage)
	if err != nil {
		return promptResponse, err
	}

	messages = m

	body := map[string]any{
		"model":                 model,
		"messages":              messages,
		"stream":                false,
		"temperature":           temperature,
		"max_completion_tokens": maxTokens,
		"response_format":       responseFormat,
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

func (c *OpenAIClient) setupSystemPromptIfNeeded(conversation ConversationRepositoryConversation, defaultPrompt string, model string) ConversationRepositoryConversation {
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

func (c *OpenAIClient) toResponseFormat(schema *map[string]any, schemaName string) *map[string]any {
	if schema == nil {
		return nil
	}

	return &map[string]any{
		"type": "json_schema",
		"json_schema": map[string]any{
			"name":   schemaName,
			"schema": schema,
		},
	}
}

func (c *OpenAIClient) writeResponseFormatTo(item *ConversationRepositoryConversationItem, schema *map[string]any, schemaName string) (*map[string]any, error) {
	responseFormat := c.toResponseFormat(schema, schemaName)
	if responseFormat != nil {
		jsonData, err := json.Marshal(responseFormat)
		if err != nil {
			return responseFormat, err
		}

		item.ResponseFormat = string(jsonData)
	}

	return responseFormat, nil
}
