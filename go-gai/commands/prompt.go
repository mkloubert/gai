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
	"errors"
	"io"
	"os"
	"strings"

	"github.com/mkloubert/gai/types"
	"github.com/spf13/cobra"
)

// Init_prompt_Command initializes the `prompt` command.
func Init_prompt_Command(app *types.AppContext, parentCmd *cobra.Command) {
	var promptCmd = &cobra.Command{
		Use:     "prompt [PROMPT]",
		Aliases: []string{"p"},
		Short:   "AI prompt",
		Long:    `Sends a prompt to AI.`,
		Run: func(cmd *cobra.Command, args []string) {
			app.InitAI()

			files, err := app.GetFiles()
			app.CheckIfError(err)

			responseSchema, responseSchemaName, err := app.GetResponseSchema()
			app.CheckIfError(err)

			prompt, err := app.GetInput(args)
			app.CheckIfError(err)

			prompt = strings.TrimSpace(prompt)
			if prompt == "" {
				app.CheckIfError(errors.New("no prompt defined"))
			}

			options := make([]types.AIClientPromptOptions, 0)

			options = append(options, types.AIClientPromptOptions{
				ResponseSchema:     responseSchema,
				ResponseSchemaName: &responseSchemaName,
			})

			openedFiles := make([]*os.File, 0)
			defer func() {
				for _, of := range openedFiles {
					of.Close()
				}
			}()

			for _, f := range files {
				file, err := os.Open(f)
				app.CheckIfError(err)

				openedFiles = append(openedFiles, file)

				options = append(options, types.AIClientPromptOptions{
					Files: &[]io.Reader{file},
				})
			}

			response, err := app.AI.Prompt(prompt, options...)
			app.CheckIfError(err)

			app.OutputAIAnswer(response.Content)
		},
	}

	app.WithPromptCLIFlags(promptCmd)

	parentCmd.AddCommand(
		promptCmd,
	)
}
