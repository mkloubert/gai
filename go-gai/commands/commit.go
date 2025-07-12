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
	"sort"
	"strings"

	"github.com/mkloubert/gai/types"
	"github.com/mkloubert/gai/utils"
	"github.com/pkoukk/tiktoken-go"
	"github.com/spf13/cobra"

	tea "github.com/charmbracelet/bubbletea"
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
	var stagedOnly bool

	var commitCmd = &cobra.Command{
		Use:   "commit",
		Short: "Commit",
		Long:  `Commits staged files with AI.`,
		Run: func(cmd *cobra.Command, args []string) {
			startTime := app.GetISOTime()

			app.Dbgf("Start time: %s%s", startTime, app.EOL)

			git, err := app.NewGitClient()
			app.CheckIfError(err)

			app.Dbg("Created new git client")

			allStagedFiles, err := git.GetStagedFiles()
			app.CheckIfError(err)

			app.Dbgf("Found %d staged files%s", len(allStagedFiles), app.EOL)

			if len(allStagedFiles) == 0 {
				// no staged files found, ask user for changed files to stage

				app.Dbg("Asking user for changed files to tage ...")

				changedFiles, err := git.GetChangedFiles()
				app.CheckIfError(err)

				app.Dbgf("Found %d changed files%s", len(changedFiles), app.EOL)

				if len(changedFiles) > 0 {
					// ask

					// but sort by name first
					sort.Slice(changedFiles, func(i, j int) bool {
						return strings.ToLower(changedFiles[i].Name()) < strings.ToLower(changedFiles[j].Name())
					})

					model := &stageChangedFilesModel{
						items:  []stageChangedFilesModelItem{},
						cursor: 0,
						done:   false,
					}

					for _, cf := range changedFiles {
						model.items = append(model.items, stageChangedFilesModelItem{
							checked: true,
							file:    cf,
							label:   cf.Name(),
						})
					}

					if !app.AlwaysYes {
						// ask user, otherwise
						// all changed are taken

						p := tea.NewProgram(model)

						_, err := p.Run()
						app.CheckIfError(err)
					} else {
						app.Dbg("Auto adding changed files ...")
					}

					for _, item := range model.items {
						if !item.checked {
							continue
						}

						app.Dbgf("Staging changed file '%s' ...%s", item.file.Name(), app.EOL)

						err := item.file.Stage()
						app.CheckIfError(err)

						allStagedFiles = append(allStagedFiles, item.file)
					}
				}
			}

			if len(allStagedFiles) == 0 {
				app.CheckIfError(errors.New("no changed or staged files found"))
			}

			app.InitAI()

			model := app.AI.ChatModel()

			app.Dbgf("Initializes AI with model %s changed files%s", model, app.EOL)

			startEmpty := true

			contextOptions := make([]types.NewChatContextOptions, 0)
			contextOptions = append(contextOptions, types.NewChatContextOptions{
				StartEmpty: &startEmpty,
			})

			chat, err := app.NewChatContext(contextOptions...)
			app.CheckIfError(err)

			app.Dbg("Created chat context")

			latestCommit, err := git.GetLatestCommit()
			app.CheckIfError(err)

			app.Dbgf("Latest git commit: %s%s", latestCommit.Hash(), app.EOL)

			allLatestCommitedFiles, err := latestCommit.GetFiles()
			app.CheckIfError(err)

			app.Dbgf("Number of files in this commit: %d%s", len(allLatestCommitedFiles), app.EOL)

			checkFile := app.NewFilePredicate()

			// before we start we filter the staged files we really want
			finalStagedFilesToTake := make([]*types.GitFile, 0)
			for _, sf := range allStagedFiles {
				takeFile, err := checkFile(sf.FullName())
				app.CheckIfError(err)

				if takeFile {
					finalStagedFilesToTake = append(finalStagedFilesToTake, sf)
				} else {
					app.Dbgf("Will not take staged file '%s'%s", sf.Name(), app.EOL)
				}
			}

			app.Dbgf("Number of final staged files to take: %d%s", len(finalStagedFilesToTake), app.EOL)

			// then we collect the latest commit files for the context ...
			finalLastCommitedFilesToTake := make([]*types.GitFile, 0)
			for _, lcf := range allLatestCommitedFiles {
				takeFile, err := checkFile(lcf.FullName())
				app.CheckIfError(err)

				if !takeFile {
					app.Dbgf("Will not take staged file '%s'%s", lcf.Name(), app.EOL)
					continue
				}

				addFile := func() {
					finalLastCommitedFilesToTake = append(finalLastCommitedFilesToTake, lcf)
				}

				if stagedOnly {
					// ... but only the one which are are part of the staged files

					for _, sf := range finalStagedFilesToTake {
						if sf.FullName() == lcf.FullName() {
							addFile()
						}
					}
				} else {
					addFile()
				}
			}

			app.Dbgf("Number of final files from latest commit to take: %d%s", len(finalLastCommitedFilesToTake), app.EOL)

			approximateSubmittedBinarySize := uint64(0)
			approximateSubmittedTextSize := uint64(0)
			approximateSubmittedText := ""

			// append files from latest commit
			app.Dbg("Appending files from latest commit ...")
			chat.AppendSimplePseudoUserConversation(`I will start by submitting each file from the latest git commit with its contents as serialized JSON strings.
Answer with 'OK' if you understand this.`,
				types.AppendSimplePseudoUserConversationOptions{
					Model: &model,
					Time:  &startTime,
				})
			for i, lcf := range finalLastCommitedFilesToTake {
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
					app.Dbgf("'%s' from latest commit seems to be binary%s", lcf.Name(), app.EOL)

					approximateSubmittedBinarySize += uint64(len(latestContent))
				} else {
					app.Dbgf("'%s' from latest commit seems to be text%s", lcf.Name(), app.EOL)

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

			app.Dbg("Appending staged files ...")
			chat.AppendSimplePseudoUserConversation(`Now I continue with the list of staged files.
Each file contains the diff (or the complete content) between the staged content and the latest commit based on the current state status.
Answer with 'OK' if you understand this.`,
				types.AppendSimplePseudoUserConversationOptions{
					Model: &model,
					Time:  &startTime,
				})
			for i, sf := range finalStagedFilesToTake {
				if app.DryRun {
					app.Writeln(fmt.Sprintf("Staged file: %s", sf.Name()))
				} else {
					app.Dbgf("Staged file: '%s'%s", sf.Name(), app.EOL)
				}

				stageStatus := sf.StageStatus()

				ensureCurrentContent := func() []byte {
					c, err := sf.GetCurrentContent()
					app.CheckIfError(err)

					if app.DryRun {
						app.Writeln(fmt.Sprintf("\tSize: %d", len(c)))
					}

					return c
				}

				messageSuffix := ""
				if i > 0 {
					messageSuffix = " and integrate it with the context of the other staged files"
				}

				app.Dbgf("'%s' from staged files has status '%s' %s", sf.Name(), stageStatus, app.EOL)

				if stageStatus == "A" {
					// added

					currentContent := ensureCurrentContent()

					var um string
					if utils.MaybeBinary(currentContent) {
						app.Dbgf("'%s' from staged files seems to be binary%s", sf.Name(), app.EOL)

						um = fmt.Sprintf(
							`This is the new binary file with the path '%s'. In this case no content is submitted.
Try to take the context from the path.
Answer with 'OK' if you analyzed it%v.`,
							sf.Name(),
							messageSuffix,
						)

						approximateSubmittedBinarySize += uint64(len(currentContent))
					} else {
						app.Dbgf("'%s' from staged files seems to be text%s", sf.Name(), app.EOL)

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

					currentContent := ensureCurrentContent()

					var um string
					if utils.MaybeBinary(currentContent) {
						// binary file
						app.Dbgf("'%s' from staged files seems to be binary%s", sf.Name(), app.EOL)

						um = fmt.Sprintf(
							`This is the updated binary file with the path '%s'. In this case no content is submitted.
Try to take the context from the path.
Answer with 'OK' if you analyzed it%v.`,
							sf.Name(),
							messageSuffix,
						)

						approximateSubmittedBinarySize += uint64(len(currentContent))
					} else {
						app.Dbgf("'%s' from staged files seems to be text%s", sf.Name(), app.EOL)

						app.Dbgf("Making diff from staged file '%s' ...%s", sf.Name(), app.EOL)

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

			responseSchema, responseSchemaName, err := app.GetResponseSchema()
			app.CheckIfError(err)

			app.Dbg("Got response schema")

			systemPrompt := fmt.Sprintf(`You are an expert assistant for writing git commit messages. Your primary goal is to generate clear, concise, and correct commit messages strictly following the [Conventional Commits](https://www.conventionalcommits.org/en/v1.0.0/) specification.

When a user describes a code change, you will:
- Analyze the description and suggest a suitable commit message.
- Always use the Conventional Commits format, including type, optional scope, and a concise description.
- If a breaking change is described, use the %s syntax or the %s footer as specified.
- When appropriate, include a commit body or footer, for example for breaking changes or references to issues.
- Format all output as plain text or Markdown, as appropriate.
- You can always assume that there is no issue behind this commit.

Below is the full Conventional Commits specification for your reference. Always follow these rules:
---

%s`, "`!`", "`BREAKING CHANGE:`",
				conventionalCommitsSpec)

			app.Dbg("Setup system prompt")

			doNotSaveConversation := true

			additionalContext, err := app.GetInput(args)
			app.CheckIfError(err)

			app.Dbg("Got additional context")

			lastMessage := "Write a commit message based on the Conventional Commits specification."
			if strings.TrimSpace(additionalContext) != "" {
				jsonData, err := json.Marshal(&additionalContext)
				app.CheckIfError(err)

				lastMessage += fmt.Sprintf(
					"\nHere is some additional context from user as serialized JSON string: %s",
					jsonData,
				)
			}
			lastMessage += "\nYour JSON:"

			app.Dbg("Setup message")

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
			} else {
				app.Dbg("Taking custom response schema")
			}
			if strings.TrimSpace(responseSchemaName) == "" {
				responseSchemaName = "CreateGitCommitMessageSchema"
			} else {
				app.Dbgf("Custom response schema name: %s%s", responseSchemaName, app.EOL)
			}

			if app.DryRun {
				app.Writeln("Stop here because of dry run mode.")

				os.Exit(0)
			}

			var nextRequest func() (string, error)
			nextRequest = func() (string, error) {
				chatOptions := make([]types.AIClientChatOptions, 0)
				chatOptions = append(chatOptions, types.AIClientChatOptions{
					NoSave:             &doNotSaveConversation,
					ResponseSchema:     responseSchema,
					ResponseSchemaName: &responseSchemaName,
					SystemPrompt:       &systemPrompt,
				})
				answer, _, err := app.AI.Chat(chat, lastMessage, chatOptions...)
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

				if doCommit {
					app.Dbg("Will do an auto commit ...")
				}

				if !doCommit {
					reader := bufio.NewReader(app.Stdin)
					for {
						app.WriteString("Commit with this message [Y(es)/r(etry)/n(no)]?: ")

						input, _ := reader.ReadString('\n')
						input = strings.TrimSpace(strings.ToLower(input))

						if input == "y" || input == "yes" || input == "" {
							doCommit = true
							break
						}
						if input == "r" || input == "retry" {
							app.WriteString("Some (optional) context for the new message: ")

							retryInput, _ := reader.ReadString('\n')
							retryInput = strings.TrimSpace(retryInput)

							if retryInput == "" {
								lastMessage = "Please create a new message."
							} else {
								lastMessage = retryInput
							}

							return nextRequest()
						}
						if input == "n" || input == "no" {
							doCommit = false
							break
						}

						app.Writeln()
						app.WriteString("Please answer with 'y' (yes), 'r' (retry) or n ('no').")
					}
				}

				if !doCommit {
					app.Writeln("Cancelled.")
					os.Exit(0)
				}

				app.Writeln()

				return finalCommitMessage, err
			}

			commitMessage, err := nextRequest()
			app.CheckIfError(err)

			app.Dbg("Running 'git commit' ...")

			c := git.CreateExecCommand("git", "commit", "-m", commitMessage)

			c.Stderr = app.Stderr
			c.Stdin = app.Stdin
			c.Stdout = app.Stdout

			err = c.Run()
			app.CheckIfError(err)
		},
	}

	app.WithDryRunCliFlags(commitCmd)
	app.WithYesCliFlags(commitCmd)
	commitCmd.Flags().BoolVarP(&stagedOnly, "staged-only", "", false, "only submit staged files for comparsion")

	parentCmd.AddCommand(
		commitCmd,
	)
}
