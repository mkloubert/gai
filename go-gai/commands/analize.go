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

package commands

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/mkloubert/gai/types"
	"github.com/spf13/cobra"
)

func init_analize_code_Command(app *types.AppContext, parentCmd *cobra.Command) {
	var analizeCodeCmd = &cobra.Command{
		Use:     "code",
		Aliases: []string{"c"},
		Short:   "Analize code",
		Long:    `Analize source code as defined in --file and --files flags.`,
		Run: func(cmd *cobra.Command, args []string) {
			startTime := app.GetISOTime()

			app.InitAI()

			files, err := app.GetFiles()
			app.CheckIfError(err)

			if len(files) == 0 {
				app.CheckIfError(errors.New("no files found or defined"))
			}

			responseSchema, responseSchemaName, err := app.GetResponseSchema()
			app.CheckIfError(err)

			message, err := app.GetInput(args)
			app.CheckIfError(err)

			message = strings.TrimSpace(message)
			if message == "" {
				app.CheckIfError(errors.New("no chat message defined"))
			}

			doNotSaveConversation := true
			startEmpty := true

			contextOptions := make([]types.NewChatContextOptions, 0)
			contextOptions = append(contextOptions, types.NewChatContextOptions{
				StartEmpty: &startEmpty,
			})

			chat, err := app.NewChatContext(contextOptions...)
			app.CheckIfError(err)

			model := app.AI.ChatModel()

			// user explaination

			chat.AppendSimplePseudoUserConversation(`I will start by submitting each file with its contents as serialized JSON strings.  
After this, I will submit my question or query, and you will follow it exactly and answer in the same language.
Answer with 'OK' if you understand this.`)

			// start creating a pseudo conversation
			chat.AppendTextFilesAsPseudoConversation(files)

			// setup final message and instructions
			{
				jsonsData, err := json.Marshal(&message)
				app.CheckIfError(err)

				userMessage := &types.ConversationRepositoryConversationItem{
					Contents: make(types.ConversationRepositoryConversationItemContents, 0),
					Model:    model,
					Role:     "user",
					Time:     startTime,
				}
				newUserTextItem := &types.ConversationRepositoryConversationItemContentItem{
					Content: fmt.Sprintf(
						`OK, this was the last file.  
Now, this is your task: %s.  
Your answer:`,
						jsonsData,
					),
					Type: "text",
				}
				userMessage.Contents = append(userMessage.Contents, newUserTextItem)

				chat.AppendConversationItem(userMessage)
			}

			// now do the chat and output anything ...

			outputLanguage := strings.TrimSpace(app.OutputLanguage)

			langInfo := "same language as input"
			if outputLanguage != "" {
				langInfo = fmt.Sprintf("'%s' language", outputLanguage)
			}

			systemPrompt := fmt.Sprintf(`You are a helpful, thorough, and articulate assistant helping a user analyze source code.
Always respond with detailed explanations, clear structure, and relevant examples.
Break down complex topics step by step.
When appropriate, use bullet points, headings, or code snippets to enhance clarity.
The user will start by submitting each file with its contents as serialized JSON strings.  
After this, the user will submit their question or query, and you will follow it exactly and answer in %s.`,
				langInfo)

			chatOptions := make([]types.AIClientChatOptions, 0)
			chatOptions = append(chatOptions, types.AIClientChatOptions{
				NoSave:             &doNotSaveConversation,
				ResponseSchema:     responseSchema,
				ResponseSchemaName: &responseSchemaName,
				SystemPrompt:       &systemPrompt,
			})
			answer, _, err := app.AI.Chat(chat, message, chatOptions...)
			app.CheckIfError(err)

			app.OutputAIAnswer(answer)
		},
	}

	app.WithChatFlags(analizeCodeCmd)
	app.WithLanguageFlags(analizeCodeCmd)

	parentCmd.AddCommand(
		analizeCodeCmd,
	)
}

// Init_analize_Command initializes the `analize` command.
func Init_analize_Command(app *types.AppContext, parentCmd *cobra.Command) {
	var analizeCmd = &cobra.Command{
		Use:     "analize [resource]",
		Aliases: []string{"a"},
		Short:   "Analize",
		Long:    `Analizes a resource.`,
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}

	init_analize_code_Command(app, analizeCmd)

	parentCmd.AddCommand(
		analizeCmd,
	)
}
