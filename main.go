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
	"fmt"
	"log"
	"os"

	"github.com/mkloubert/gai/commands"
	"github.com/mkloubert/gai/types"

	"github.com/spf13/cobra"
)

func main() {
	var app *types.AppContext

	rootCmd := &cobra.Command{
		Use:   "gai",
		Short: "gAI is a command line tool for AI tasks",
		Long:  "A command line to for AI tasks which can be found at https://github.com/mkloubert/gai.",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			app.Init()

			app.Dbg(fmt.Sprintf("Executing command '%v' ...", cmd.Name()))
		},
		Run: func(cmd *cobra.Command, args []string) {
			commands.RunRootCommand(app, cmd, args)
		},
	}

	// setup app command
	app = &types.AppContext{
		RootCommand: rootCmd,
		Stderr:      os.Stderr,
		Stdin:       os.Stdin,
		Stdout:      os.Stdout,
	}

	// use these flags and options everywhere
	rootCmd.PersistentFlags().StringVarP(&app.WorkingDirectory, "cwd", "", "", "current working directory")
	rootCmd.PersistentFlags().StringVarP(&app.EOL, "eol", "", fmt.Sprintln(), "custom EOL char sequence")
	rootCmd.PersistentFlags().StringArrayVarP(&app.EnvFiles, "env-file", "", []string{}, "one or more env file to load")
	rootCmd.PersistentFlags().StringVarP(&app.HomeDirectory, "home", "", "", "user's home directory")
	rootCmd.PersistentFlags().BoolVarP(&app.SkipDefaultEnvFiles, "skip-env-files", "", false, "do not load default .env files")
	rootCmd.PersistentFlags().StringVarP(&app.Model, "model", "", "", "default chat model")
	rootCmd.PersistentFlags().BoolVarP(&app.Verbose, "verbose", "", false, "verbose output")

	commands.Init_ask_Command(app, rootCmd)
	commands.Init_list_Command(app, rootCmd)

	app.Log = log.New(app, "", log.Ldate|log.Ltime)

	// execute app
	app.Run()
}
