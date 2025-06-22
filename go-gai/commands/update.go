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
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/mkloubert/gai/types"
	"github.com/spf13/cobra"
)

type updateCodeResponse struct {
	UpdatedFiles map[string]updateCodeResponseFileToUpdateToUpdate `json:"updated_files"`
}

type updateCodeResponseFileToUpdateToUpdate struct {
	Explanation string `json:"explanation"`
	NewContent  string `json:"new_content"`
}

func init_update_code_Command(app *types.AppContext, parentCmd *cobra.Command) {
	var updateCodeCmd = &cobra.Command{
		Use:     "code",
		Aliases: []string{"c"},
		Short:   "Update code",
		Long:    `Updates source code as defined in --file and --files flags.`,
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
Answer with 'OK' if you understand this.`,
				types.AppendSimplePseudoUserConversationOptions{
					Model: &model,
					Time:  &startTime,
				},
			)

			filesToUpdate := make([]string, 0)

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
Your JSON:`,
						jsonsData,
					),
					Type: "text",
				}
				userMessage.Contents = append(userMessage.Contents, newUserTextItem)

				chat.AppendConversationItem(userMessage)
			}

			if responseSchema == nil {
				// build default schema

				required1 := make([]string, 0)
				properties1 := map[string]any{}

				for _, f := range filesToUpdate {
					required1 = append(required1, f)

					properties1[f] = map[string]any{
						"description": fmt.Sprintf("Information how the file '%v' should be updated.", f),
						"type":        "object",
						"required":    []string{"explanation", "new_content"},
						"properties": map[string]any{
							"explanation": map[string]any{
								"type":        "string",
								"description": fmt.Sprintf("Detailed explanation of what has been changed in file '%s'.", f),
							},
							"new_content": map[string]any{
								"type":        "string",
								"description": fmt.Sprintf("New content for file '%s'.", f),
							},
						},
					}
				}

				responseSchema = &map[string]any{
					"type":     "object",
					"required": []string{"updated_files"},
					"properties": map[string]any{
						"updated_files": map[string]any{
							"type":        "object",
							"description": "List of files that should be updated.",
							"required":    required1,
							"properties":  properties1,
						},
					},
				}
			}
			if strings.TrimSpace(responseSchemaName) == "" {
				responseSchemaName = "FileUpdateSchema"
			}

			outputLanguage := strings.TrimSpace(app.OutputLanguage)

			langInfo := "same language as input"
			if outputLanguage != "" {
				langInfo = fmt.Sprintf("'%s' language", outputLanguage)
			}

			systemPrompt := fmt.Sprintf(`You are a skilled and helpful software developer acting as a code reviewer. Your job is to help the user analyze, update, and improve their source code.
The user will submit each file as a JSON string containing its filename and contents.
After all relevant files have been submitted, the user will send a question or request related to the codebase.
Please follow the user’s instructions precisely and answer in %s.`,
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

			var updateResponse updateCodeResponse
			err = json.Unmarshal([]byte(answer), &updateResponse)
			app.CheckIfError(err)

			for fileName, fileItem := range updateResponse.UpdatedFiles {
				if !slices.Contains(filesToUpdate, fileName) {
					app.CheckIfError(fmt.Errorf("%s is an unknown file that cannot be updated", fileName))
				}

				fullPath := filepath.Join(app.WorkingDirectory, fileName)

				stat, err := os.Stat(fullPath)
				app.CheckIfError(err)

				err = os.WriteFile(fullPath, []byte(fileItem.NewContent), stat.Mode().Perm())
				app.CheckIfError(err)

				app.OutputAIAnswer(fmt.Sprintf(
					`Updated *%s*:
%s%s`,
					fileName,
					fileItem.Explanation,
					app.EOL,
				))
			}
		},
	}

	app.WithChatFlags(updateCodeCmd)
	app.WithLanguageFlags(updateCodeCmd)

	parentCmd.AddCommand(
		updateCodeCmd,
	)
}

// Init_update_Command initializes the `update` command.
func Init_update_Command(app *types.AppContext, parentCmd *cobra.Command) {
	var listCmd = &cobra.Command{
		Use:     "update [resource]",
		Aliases: []string{"u"},
		Short:   "Update",
		Long:    `Updates a resource.`,
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}

	init_update_code_Command(app, listCmd)

	parentCmd.AddCommand(
		listCmd,
	)
}
