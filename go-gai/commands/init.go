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
	"strings"

	"github.com/gosimple/slug"
	"github.com/mkloubert/gai/types"
	"github.com/mkloubert/gai/utils"
	"github.com/spf13/cobra"
)

type initCodeResponseProjectFile struct {
	Explanation      string `json:"explanation"`
	RelativeFilePath string `json:"relative_file_path"`
	TextContent      string `json:"text_content"`
}

type initCodeResponseProject struct {
	ProjectFiles []initCodeResponseProjectFile `json:"project_files"`
	Readme       string                        `json:"readme"`
}

const allowedFileChars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ_-.0123456789/ "

func cleanupPath(s string) string {
	s = strings.ReplaceAll(s, fmt.Sprintf("%c", os.PathSeparator), "/")
	s = strings.ReplaceAll(s, "\t", "  ")
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.ReplaceAll(s, "\r", " ")

	result := ""

	for _, c := range s {
		char := fmt.Sprintf("%c", c)

		if strings.Contains(allowedFileChars, char) {
			result = result + char
		} else {
			result = result + "_"
		}
	}

	return strings.TrimSpace(result)
}

func init_init_project_Command(app *types.AppContext, parentCmd *cobra.Command) {
	var initCodeCmd = &cobra.Command{
		Use:     "code [project]",
		Args:    cobra.MinimumNArgs(1),
		Aliases: []string{"c"},
		Short:   "Init code",
		Long:    `Initializes source code project.`,
		Run: func(cmd *cobra.Command, args []string) {
			app.InitAI()

			responseSchema, responseSchemaName, err := app.GetResponseSchema()
			app.CheckIfError(err)

			projectName := strings.TrimSpace(slug.Make(args[0]))
			if projectName == "" {
				app.CheckIfError(errors.New("no valid project name provided"))
			}

			projectRoot := filepath.Join(app.WorkingDirectory, projectName)

			info, err := os.Stat(projectRoot)
			if err == nil {
				app.CheckIfError(fmt.Errorf("'%s' already exists", info.Name()))
			}

			err = os.MkdirAll(projectRoot, 0755)
			app.CheckIfError(err)

			message, err := app.GetInput(args[1:])
			app.CheckIfError(err)

			outputLanguage := strings.TrimSpace(app.OutputLanguage)

			langInfo := "same language as input"
			if outputLanguage != "" {
				langInfo = fmt.Sprintf("'%s' language", outputLanguage)
			}

			systemPrompt := fmt.Sprintf(`You are an expert in setting up software projects.
Follow the user’s instructions precisely and completely.
Provide a clear, structured, and fully detailed response with all necessary steps, commands, code examples, or configuration files, so that the user can proceed without needing any further clarification.
Distribute the project to several files and subfolders if necessary and useful.
Create as many files and directories as possible with all required content so that the user has to do as little rework as possible and can get started quickly.
Always answer in %s.`,
				langInfo,
			)

			if responseSchema == nil {
				responseSchema = &map[string]any{
					"type":     "object",
					"required": []string{"project_files", "readme"},
					"properties": map[string]any{
						"project_files": map[string]any{
							"type":        "array",
							"description": "List of files to create.",
							"items": map[string]any{
								"type":     "object",
								"required": []string{"text_content", "explanation", "relative_file_path"},
								"properties": map[string]any{
									"relative_file_path": map[string]any{
										"description": "Relative path of the file to create using / as path separators.",
										"type":        "string",
									},
									"text_content": map[string]any{
										"description": "Text content or data URI if file contains binary data like image, audio or video.",
										"type":        "string",
									},
									"explanation": map[string]any{
										"description": "Detailed information what the file is and for what it is used for.",
										"type":        "string",
									},
								},
							},
						},
						"readme": map[string]any{
							"description": "Markdown with detailed information on what is required to start the project.",
							"type":        "string",
						},
					},
				}

			}
			if strings.TrimSpace(responseSchemaName) == "" {
				responseSchemaName = "CreateSoftwareProjectSchema"
			}

			promptOptions := make([]types.AIClientPromptOptions, 0)
			promptOptions = append(promptOptions, types.AIClientPromptOptions{
				ResponseSchema:     responseSchema,
				ResponseSchemaName: &responseSchemaName,
				SystemPrompt:       &systemPrompt,
			})
			response, err := app.AI.Prompt(message, promptOptions...)
			app.CheckIfError(err)

			var newProject initCodeResponseProject
			err = json.Unmarshal([]byte(response.Content), &newProject)
			app.CheckIfError(err)

			ensureDirOfFile := func(fp string) error {
				dir := filepath.Dir(fp)
				err := os.MkdirAll(dir, 0755)

				return err
			}

			for _, newFile := range newProject.ProjectFiles {
				relPath := cleanupPath(newFile.RelativeFilePath)
				fullPath := filepath.Join(projectRoot, relPath)

				requiredPrefix := fmt.Sprintf("%s%c", projectRoot, os.PathSeparator)
				if !strings.HasPrefix(fullPath, requiredPrefix) {
					app.CheckIfError(fmt.Errorf("invalid file path '%s'", fullPath))
				}

				err := ensureDirOfFile(fullPath)
				app.CheckIfError(err)

				var data []byte

				dataUri := strings.TrimSpace(newFile.TextContent)
				if strings.HasPrefix(dataUri, "data:") {
					d, err := utils.DataURIToBytes(dataUri)
					if err != nil {
						d = []byte(dataUri) // fallback
					}

					data = d
				} else {
					data = []byte(dataUri)
				}

				err = os.WriteFile(fullPath, data, 0644)
				app.CheckIfError(err)

				app.OutputAIAnswer(fmt.Sprintf(
					`Created *%s*:
%s%s`,
					relPath,
					newFile.Explanation,
					app.EOL,
				))
			}

			// README file
			{
				relPath := "README.md"
				readmeFile := filepath.Join(projectRoot, relPath)

				err := ensureDirOfFile(readmeFile)
				app.CheckIfError(err)

				err = os.WriteFile(readmeFile, []byte(newProject.Readme), 0644)
				app.CheckIfError(err)

				app.OutputAIAnswer(fmt.Sprintf(
					`Finally created *%s*%s`,
					relPath,
					app.EOL,
				))
			}
		},
	}

	app.WithLanguageFlags(initCodeCmd)

	parentCmd.AddCommand(
		initCodeCmd,
	)
}

// Init_init_Command initializes the `init` command.
func Init_init_Command(app *types.AppContext, parentCmd *cobra.Command) {
	var initCmd = &cobra.Command{
		Use:     "init [resource]",
		Aliases: []string{"init", "initialize", "i"},
		Short:   "Initialize",
		Long:    `Initializes a resource.`,
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}

	init_init_project_Command(app, initCmd)

	parentCmd.AddCommand(
		initCmd,
	)
}
