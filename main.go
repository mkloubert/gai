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

package main

import (
	"os"

	"github.com/mkloubert/gai/commands"
	"github.com/mkloubert/gai/types"

	"github.com/mkloubert/gai/utils"
	"github.com/spf13/cobra"
)

func main() {
	// current working directory
	cwd, err := os.Getwd()
	utils.CheckIfError(err)

	var app *types.AppContext

	// setup app command
	app = &types.AppContext{
		RootCommand: &cobra.Command{
			Use:   "gai",
			Short: "gai is a command line tool for AI tasks",
			Long:  "A command line to for AI tasks which can be found at https://github.com/mkloubert/gai.",
			Run: func(cmd *cobra.Command, args []string) {
				commands.RunRootCommand(app, cmd, args)
			},
		},
		Stderr:           os.Stderr,
		WorkingDirectory: cwd,
	}

	// load .env file, if exists
	app.LoadEnvFilesIfExist()

	// execute app
	app.Run()
}
