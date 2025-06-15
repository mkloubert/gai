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
	"strings"

	"github.com/mkloubert/gai/types"
	"github.com/spf13/cobra"
)

// Init_chat_Command initializes the `chat` command.
func Init_chat_Command(app *types.AppContext, parentCmd *cobra.Command) {
	var reset bool

	var chatCmd = &cobra.Command{
		Use:     "chat [QUESTION]",
		Aliases: []string{"c"},
		Short:   "AI chat",
		Long:    `Asks the AI a question.`,
		Run: func(cmd *cobra.Command, args []string) {
			message := strings.TrimSpace(
				strings.Join(args, " "),
			)
			if message == "" {
				app.CheckIfError(errors.New("no chat message defined"))
			}

			chat, err := app.NewChatContext()
			app.CheckIfError(err)

			if reset {
				chat.ResetConversation()
			}

			answer, _, err := app.AI.Chat(chat, message)
			app.CheckIfError(err)

			app.Writeln(answer)

			err = chat.UpdateConversation()
			app.CheckIfError(err)
		},
	}

	chatCmd.Flags().BoolVarP(&reset, "reset", "r", false, "reset conversation")

	parentCmd.AddCommand(
		chatCmd,
	)
}
