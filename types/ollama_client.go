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
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// OllamaClient is an `AIClient` implementation for Ollama.
type OllamaClient struct {
	chatModel string
}

// Chat starts or continues a chat conversation with message in `msg` based on `ctx` and returns the new conversation.
func (c *OllamaClient) Chat(ctx *ChatContext, msg string) (string, ConversationRepositoryConversation, error) {
	conversation, err := ctx.GetConversation()
	if err != nil {
		return "", conversation, err
	}

	model := strings.TrimSpace(strings.ToLower(c.chatModel))
	if model == "" {
		return "", conversation, fmt.Errorf("no chat ai model defined")
	}

	url := "http://localhost:11434/api/chat"

	userMessage := OllamaAIChatMessage{
		Content: msg,
		Role:    "user",
	}

	messages := []OllamaAIChatMessage{}
	// add previuos conversation
	for _, item := range conversation {
		messages = append(messages, OllamaAIChatMessage{
			Content: item.Content,
			Role:    item.Role,
		})
	}
	// add user message
	messages = append(messages, userMessage)

	body := map[string]interface{}{
		"model":    c.chatModel,
		"messages": messages,
		"stream":   false,
		// TODO: read from settings
		"temperature": 0.3,
	}

	jsonData, err := json.Marshal(&body)
	if err != nil {
		return "", conversation, err
	}

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

	if resp.StatusCode != 200 {
		return "", conversation, fmt.Errorf("unexpected response %v", resp.StatusCode)
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

	assistantMessage := OpenAIChatMessage{
		Content: chatResponse.Message.Content,
		Role:    "assistant",
	}

	answer := assistantMessage.Content

	conversation = append(conversation, &ConversationRepositoryConversationItem{
		Role:    "user",
		Content: userMessage.Content,
	})
	conversation = append(conversation, &ConversationRepositoryConversationItem{
		Role:    "assistant",
		Content: answer,
	})

	ctx.UpdateConversationWith(conversation)

	return answer, conversation, nil
}

// ChatModel returns the current chat model.
func (c *OllamaClient) ChatModel() string {
	return c.chatModel
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
