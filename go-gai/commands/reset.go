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
	"github.com/mkloubert/gai/types"
	"github.com/spf13/cobra"
)

func init_reset_conversation_Command(app *types.AppContext, parentCmd *cobra.Command) {
	var resetConversationCmd = &cobra.Command{
		Use:     "conversation",
		Aliases: []string{"c"},
		Short:   "Reset conversation",
		Long:    `Reset conversation of current context.`,
		Run: func(cmd *cobra.Command, args []string) {
			chat, err := app.NewChatContext()
			app.CheckIfError(err)

			err = chat.ResetConversation()
			app.CheckIfError(err)

			err = chat.UpdateConversation()
			app.CheckIfError(err)
		},
	}

	parentCmd.AddCommand(
		resetConversationCmd,
	)
}

// Init_reset_Command initializes the `reset` command.
func Init_reset_Command(app *types.AppContext, parentCmd *cobra.Command) {
	var resetCmd = &cobra.Command{
		Use:     "reset [resource]",
		Aliases: []string{"r"},
		Short:   "Reset",
		Long:    `Resets a resource.`,
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}

	init_reset_conversation_Command(app, resetCmd)

	parentCmd.AddCommand(
		resetCmd,
	)
}
