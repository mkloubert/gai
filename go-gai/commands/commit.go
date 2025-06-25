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
	"bufio"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/mkloubert/gai/types"
	"github.com/mkloubert/gai/utils"
	"github.com/pkoukk/tiktoken-go"
	"github.com/spf13/cobra"
)

type commitResponse struct {
	Body        *string `json:"body,omitempty"`
	Description string  `json:"description"`
	Footer      *string `json:"footer,omitempty"`
	Scope       *string `json:"scope,omitempty"`
	Type        string  `json:"type"`
}

//go:embed res/conventional-commits/index.md
var conventionalCommitsSpec string

// Init_commit_Command initializes the `chat` command.
func Init_commit_Command(app *types.AppContext, parentCmd *cobra.Command) {
	var commitCmd = &cobra.Command{
		Use:   "commit",
		Short: "Commit",
		Long:  `Commits staged files with AI.`,
		Run: func(cmd *cobra.Command, args []string) {
			startTime := app.GetISOTime()

			git, err := app.NewGitClient()
			app.CheckIfError(err)

			stagedFiles, err := git.GetStagedFiles()
			app.CheckIfError(err)

			app.InitAI()

			model := app.AI.ChatModel()

			startEmpty := true

			contextOptions := make([]types.NewChatContextOptions, 0)
			contextOptions = append(contextOptions, types.NewChatContextOptions{
				StartEmpty: &startEmpty,
			})

			chat, err := app.NewChatContext(contextOptions...)
			app.CheckIfError(err)

			latestCommit, err := git.GetLatestCommit()
			app.CheckIfError(err)

			latestCommitedFiles, err := latestCommit.GetFiles()
			app.CheckIfError(err)

			checkFile := app.NewFilePredicate()

			approximateSubmittedBinarySize := uint64(0)
			approximateSubmittedTextSize := uint64(0)
			approximateSubmittedText := ""

			// append files from latest commit
			chat.AppendSimplePseudoUserConversation(`I will start by submitting each file from the latest git commit with its contents as serialized JSON strings.
Answer with 'OK' if you understand this.`,
				types.AppendSimplePseudoUserConversationOptions{
					Model: &model,
					Time:  &startTime,
				})
			for i, lcf := range latestCommitedFiles {
				takeFile, err := checkFile(lcf.FullName())
				app.CheckIfError(err)

				if !takeFile {
					continue
				}

				if app.DryRun {
					app.Writeln(fmt.Sprintf("File from last commit: %s", lcf.Name()))
				}

				latestContent, err := lcf.GetLatestContent()
				app.CheckIfError(err)

				if app.DryRun {
					app.Writeln(fmt.Sprintf("\tSize: %d", len(latestContent)))
				}

				messageSuffix := ""
				if i > 0 {
					messageSuffix = " and integrate it with the context of the other files from latest git commit"
				}

				if utils.MaybeBinary(latestContent) {
					approximateSubmittedBinarySize += uint64(len(latestContent))
				} else {
					textContent, err := utils.EnsurePlainText(latestContent)
					app.CheckIfError(err)

					jsonData, err := json.Marshal(&textContent)
					app.CheckIfError(err)

					str := string(jsonData)

					approximateSubmittedTextSize += uint64(len(jsonData))
					approximateSubmittedText += str

					chat.AppendSimplePseudoUserConversation(fmt.Sprintf(
						`This is the content of the file with the path '%s' from latest git commit: %s.
Answer with 'OK' if you analyzed it%v.`,
						lcf.Name(),
						str,
						messageSuffix,
					),
						types.AppendSimplePseudoUserConversationOptions{
							Model: &model,
							Time:  &startTime,
						},
					)
				}
			}

			stagedFilesTaken := 0

			chat.AppendSimplePseudoUserConversation(`Now I continue with the list of staged files.
Each file contains the diff (or the complete content) between the staged content and the latest commit based on the current state status.
Answer with 'OK' if you understand this.`,
				types.AppendSimplePseudoUserConversationOptions{
					Model: &model,
					Time:  &startTime,
				})
			for i, sf := range stagedFiles {
				takeFile, err := checkFile(sf.FullName())
				app.CheckIfError(err)

				if !takeFile {
					continue
				}

				if app.DryRun {
					app.Writeln(fmt.Sprintf("Staged file: %s", sf.Name()))
				}

				stagedFilesTaken += 1

				currentContent, err := sf.GetCurrentContent()
				app.CheckIfError(err)

				if app.DryRun {
					app.Writeln(fmt.Sprintf("\tSize: %d", len(currentContent)))
				}

				messageSuffix := ""
				if i > 0 {
					messageSuffix = " and integrate it with the context of the other staged files"
				}

				stageStatus := sf.StageStatus()
				if stageStatus == "A" {
					// added

					var um string
					if utils.MaybeBinary(currentContent) {
						um = fmt.Sprintf(
							`This is the new binary file with the path '%s'. In this case no content is submitted.
Try to take the context from the path.
Answer with 'OK' if you analyzed it%v.`,
							sf.Name(),
							messageSuffix,
						)

						approximateSubmittedBinarySize += uint64(len(currentContent))
					} else {
						str, err := utils.EnsurePlainText(currentContent)
						app.CheckIfError(err)

						approximateSubmittedTextSize += uint64(len(currentContent))
						approximateSubmittedText += str

						jsonData, err := json.Marshal(&str)
						app.CheckIfError(err)

						um = fmt.Sprintf(
							`This is the complete text content of the new file with the path '%s': %s.
Answer with 'OK' if you analyzed it%v.`,
							sf.Name(),
							jsonData,
							messageSuffix,
						)
					}

					chat.AppendSimplePseudoUserConversation(um,
						types.AppendSimplePseudoUserConversationOptions{
							Model: &model,
							Time:  &startTime,
						},
					)
				} else if stageStatus == "M" {
					// modified

					var um string
					if utils.MaybeBinary(currentContent) {
						// binary file

						um = fmt.Sprintf(
							`This is the updated binary file with the path '%s'. In this case no content is submitted.
Try to take the context from the path.
Answer with 'OK' if you analyzed it%v.`,
							sf.Name(),
							messageSuffix,
						)

						approximateSubmittedBinarySize += uint64(len(currentContent))
					} else {
						diff, err := latestCommit.Diff(sf)
						app.CheckIfError(err)

						approximateSubmittedTextSize += uint64(len([]byte(diff)))
						approximateSubmittedText += diff

						jsonData, err := json.Marshal(&diff)
						app.CheckIfError(err)

						um = fmt.Sprintf(
							`This is the diff content for the updated file with the path '%s': %s.
Answer with 'OK' if you analyzed it%v.`,
							sf.Name(),
							jsonData,
							messageSuffix,
						)
					}

					chat.AppendSimplePseudoUserConversation(um,
						types.AppendSimplePseudoUserConversationOptions{
							Model: &model,
							Time:  &startTime,
						},
					)
				} else if stageStatus == "D" {
					// deleted

					// user message with file and content
					chat.AppendSimplePseudoUserConversation(fmt.Sprintf(
						`The file with the path '%s' has been deleted.
Answer with 'OK' if you analyzed it%v.`,
						sf.Name(),
						messageSuffix,
					),
						types.AppendSimplePseudoUserConversationOptions{
							Model: &model,
							Time:  &startTime,
						},
					)
				}
			}

			if app.DryRun {
				app.Writeln(fmt.Sprintf("Approximate size of the total text content transferred: %d", approximateSubmittedTextSize))
				app.Writeln(fmt.Sprintf("Approximate size of the total binary content transferred: %d", approximateSubmittedBinarySize))

				// tokens
				{
					tkm, err := tiktoken.EncodingForModel(model)
					if err != nil {
						app.Writeln(fmt.Sprintf("WARN: Could not get GPT tokens for text content transfered: %s", err.Error()))
					} else {
						tokens := tkm.Encode(approximateSubmittedText, nil, nil)

						app.Writeln(fmt.Sprintf("Approximate GPT tokens for text content transfered: %d", len(tokens)))
					}
				}
			}

			if stagedFilesTaken == 0 {
				app.CheckIfError(errors.New("no files staged files found"))
			}

			responseSchema, responseSchemaName, err := app.GetResponseSchema()
			app.CheckIfError(err)

			systemPrompt := fmt.Sprintf(`You are an expert assistant for writing git commit messages. Your primary goal is to generate clear, concise, and correct commit messages strictly following the [Conventional Commits](https://www.conventionalcommits.org/en/v1.0.0/) specification.

When a user describes a code change, you will:
- Analyze the description and suggest a suitable commit message.
- Always use the Conventional Commits format, including type, optional scope, and a concise description.
- If a breaking change is described, use the %s syntax or the %s footer as specified.
- When appropriate, include a commit body or footer, for example for breaking changes or references to issues.
- Format all output as plain text or Markdown, as appropriate.

If you are unsure about the type, clearly state your assumptions or ask the user for clarification.

Below is the full Conventional Commits specification for your reference. Always follow these rules:
---

%s`, "`!`", "`BREAKING CHANGE:`",
				conventionalCommitsSpec)

			doNotSaveConversation := true

			additionalContext, err := app.GetInput(args)
			app.CheckIfError(err)

			message := "Write a commit message based on the Conventional Commits specification."
			if strings.TrimSpace(additionalContext) != "" {
				jsonData, err := json.Marshal(&additionalContext)
				app.CheckIfError(err)

				message += fmt.Sprintf(
					"\nHere is some additional context from user as serialized JSON string: %s",
					jsonData,
				)
			}
			message += "\nYour JSON:"

			if responseSchema == nil {
				responseSchema = &map[string]any{
					"type":        "object",
					"required":    []string{"description", "type"},
					"description": "Information about a commit message based on the Conventional Commits specification.",
					"properties": map[string]any{
						"type": map[string]any{
							"type":        "string",
							"description": "The type of the commit.",
						},
						"scope": map[string]any{
							"type":        "string",
							"description": "The optional scope.",
						},
						"description": map[string]any{
							"type":        "string",
							"description": "The description.",
						},
						"body": map[string]any{
							"type":        "string",
							"description": "The optional body text.",
						},
						"footer": map[string]any{
							"type":        "string",
							"description": "The option footer text(s).",
						},
					},
				}
			}
			if strings.TrimSpace(responseSchemaName) == "" {
				responseSchemaName = "CreateGitCommitMessageSchema"
			}

			if app.DryRun {
				app.Writeln("Stop here because of dry run mode.")

				os.Exit(0)
			}

			chatOptions := make([]types.AIClientChatOptions, 0)
			chatOptions = append(chatOptions, types.AIClientChatOptions{
				NoSave:             &doNotSaveConversation,
				ResponseSchema:     responseSchema,
				ResponseSchemaName: &responseSchemaName,
				SystemPrompt:       &systemPrompt,
			})
			answer, _, err := app.AI.Chat(chat, message, chatOptions...)
			app.CheckIfError(err)

			var commit commitResponse
			err = json.Unmarshal([]byte(answer), &commit)
			app.CheckIfError(err)

			commitType := strings.TrimSpace(commit.Type)
			commitScope := ""
			if commit.Scope != nil {
				commitScope = strings.TrimSpace(*commit.Scope)
			}
			commitDescription := strings.TrimSpace(commit.Description)
			commitBody := ""
			if commit.Body != nil {
				commitBody = strings.TrimSpace(*commit.Body)
			}
			commitFooter := ""
			if commit.Footer != nil {
				commitFooter = strings.TrimSpace(*commit.Footer)
			}

			if commitScope != "" {
				commitScope = fmt.Sprintf("(%s)", commitScope)
			}

			finalCommitMessage := strings.TrimSpace(
				fmt.Sprintf(`%s%s: %s

%s

%s`,
					commitType,
					commitScope,
					commitDescription,
					commitBody,
					commitFooter,
				),
			)

			app.Writeln(finalCommitMessage)
			app.Writeln()

			doCommit := app.AlwaysYes
			if !doCommit {
				reader := bufio.NewReader(app.Stdin)
				for {
					app.WriteString("Commit with this message (Y/n)?: ")

					input, _ := reader.ReadString('\n')
					input = strings.TrimSpace(strings.ToLower(input))

					if input == "y" || input == "yes" || input == "" {
						doCommit = true
						break
					}
					if input == "n" || input == "no" {
						doCommit = false
						break
					}

					app.WriteString("Please answer with 'y' (yes) or n ('no').")
				}
			}

			if !doCommit {
				app.Writeln("Cancelled.")
				os.Exit(0)
			}

			app.Writeln()

			c := exec.Command("git", "commit", "-m", finalCommitMessage)
			c.Dir = git.Dir()

			c.Stderr = app.Stderr
			c.Stdin = app.Stdin
			c.Stdout = app.Stdout

			err = c.Run()
			app.CheckIfError(err)
		},
	}

	app.WithDryRunCliFlags(commitCmd)
	app.WithYesCliFlags(commitCmd)

	parentCmd.AddCommand(
		commitCmd,
	)
}
