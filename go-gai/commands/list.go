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
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/joho/godotenv"
	"github.com/mkloubert/gai/types"
	"github.com/mkloubert/gai/utils"
	"github.com/spf13/cobra"
)

func init_list_conversation_Command(app *types.AppContext, parentCmd *cobra.Command) {
	var listConversationCmd = &cobra.Command{
		Use:     "conversation",
		Aliases: []string{"c"},
		Short:   "List conversation",
		Long:    `List conversation of current context.`,
		Run: func(cmd *cobra.Command, args []string) {
			chat, err := app.NewChatContext()
			app.CheckIfError(err)

			conversation, err := chat.GetConversation()
			app.CheckIfError(err)

			chroma := app.GetChromaSettings()

			for i, item := range conversation {
				if i > 0 {
					app.Writeln()
				}

				app.Writeln(fmt.Sprintf("%v:", item.Role))
				for _, content := range item.Contents {
					chroma.HighlightMarkdown(content.Content)
					app.Writeln()
				}
			}
		},
	}

	parentCmd.AddCommand(
		listConversationCmd,
	)
}

func init_list_env_Command(app *types.AppContext, parentCmd *cobra.Command) {
	var noSort bool

	var listEnvCmd = &cobra.Command{
		Use:     "env",
		Aliases: []string{"e"},
		Short:   "List env vars",
		Long:    `Lists all environment variables for the app in an ordered list.`,
		Run: func(cmd *cobra.Command, args []string) {
			keys := make([]string, 0)
			for key := range app.EnvVars {
				keys = append(keys, key)
			}

			app.Dbg(fmt.Sprintf("Found %v keys", len(keys)))

			if !noSort {
				app.Dbg("Sorting keys ...")

				sort.Strings(keys)
			}

			output, err := godotenv.Marshal(app.EnvVars)
			app.CheckIfError(err)

			app.WriteString(output)
		},
	}

	listEnvCmd.Flags().BoolVarP(&noSort, "no-sort", "", false, "do not sort environment variables")

	parentCmd.AddCommand(
		listEnvCmd,
	)
}

func init_list_files_Command(app *types.AppContext, parentCmd *cobra.Command) {
	var fullPath bool
	var withTypes bool

	var listFilesCmd = &cobra.Command{
		Use:   "files",
		Short: "List files",
		Long:  `Lists files as defined in --file and --files flags.`,
		Run: func(cmd *cobra.Command, args []string) {
			files, err := app.GetFiles()
			app.CheckIfError(err)

			for i, f := range files {
				if i > 0 {
					app.Writeln()
				}

				if fullPath {
					app.WriteString(f)
				} else {
					relPath, err := filepath.Rel(app.WorkingDirectory, f)
					if err != nil {
						app.WriteErrorString(fmt.Sprintf("WARN: %s%s", err.Error(), app.EOL))

						app.WriteString(f)
					} else {
						app.WriteString(relPath)
					}
				}

				if withTypes {
					data, err := os.ReadFile(f)
					if err != nil {
						app.WriteErrorString(fmt.Sprintf("ERROR: %s", err.Error()))
					} else {
						mimeType := utils.DetectMime(data)

						app.WriteString("\t")
						app.WriteString(mimeType)
					}
				}
			}
		},
	}

	listFilesCmd.Flags().BoolVarP(&fullPath, "full", "", false, "full path")
	listFilesCmd.Flags().BoolVarP(&withTypes, "with-types", "", false, "with mime types")

	parentCmd.AddCommand(
		listFilesCmd,
	)
}

func init_list_models_Command(app *types.AppContext, parentCmd *cobra.Command) {
	var listFilesCmd = &cobra.Command{
		Use:   "models",
		Short: "List models",
		Long:  `Lists AI models for each supported provider.`,
		Run: func(cmd *cobra.Command, args []string) {
			clients := make([]types.AIClient, 0)

			ollama, ollamaErr := app.NewAIClient("ollama")
			if ollamaErr != nil {
				app.Dbgf("WARN: could not create Ollama client: %s%s", ollamaErr.Error(), app.EOL)
			} else {
				clients = append(clients, ollama)
			}

			openai, openaiErr := app.NewAIClient("openai")
			if openaiErr != nil {
				app.Dbgf("WARN: could not create OpenAI client: %s%s", openaiErr.Error(), app.EOL)
			} else {
				clients = append(clients, openai)
			}

			modelList := make([]types.AIModel, 0)

			for _, c := range clients {
				loadedModels, err := c.GetModels()
				if err != nil {
					app.Dbgf("WARN: Could not load models: %s", err.Error())
					continue
				}

				modelList = append(modelList, loadedModels...)
			}

			sort.Slice(modelList, func(x, y int) bool {
				strX := modelList[x].String()
				strY := modelList[x].String()

				return strings.TrimSpace(
					strings.ToLower(strX),
				) < strings.TrimSpace(
					strings.ToLower(strY),
				)
			})

			for _, m := range modelList {
				app.Writeln(m.String())
			}
		},
	}

	parentCmd.AddCommand(
		listFilesCmd,
	)
}

// Init_list_Command initializes the `list` command.
func Init_list_Command(app *types.AppContext, parentCmd *cobra.Command) {
	var listCmd = &cobra.Command{
		Use:     "list [resource]",
		Aliases: []string{"l"},
		Short:   "List",
		Long:    `Lists a resource.`,
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}

	init_list_conversation_Command(app, listCmd)
	init_list_env_Command(app, listCmd)
	init_list_files_Command(app, listCmd)
	init_list_models_Command(app, listCmd)

	parentCmd.AddCommand(
		listCmd,
	)
}
