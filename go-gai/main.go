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
	app := &types.AppContext{
		Stderr: os.Stderr,
		Stdin:  os.Stdin,
		Stdout: os.Stdout,
	}

	rootCmd := &cobra.Command{
		Use:   "gai",
		Short: "gAI is a command line tool for AI tasks",
		Long:  "A command line to for AI tasks which can be found at https://github.com/mkloubert/gai",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			app.Init()
			app.Dbg(fmt.Sprintf("Executing command '%v' ...", cmd.Name()))
		},
		Run: func(cmd *cobra.Command, args []string) {
			commands.RunRootCommand(app, cmd, args)
		},
	}

	app.RootCommand = rootCmd

	// Define persistent flags
	flags := rootCmd.PersistentFlags()
	flags.StringVarP(&app.ApiKey, "api-key", "k", "", "global API key to use")
	flags.StringVarP(&app.BaseUrl, "base-url", "u", "", "custom base URL")
	flags.StringVarP(&app.Context, "context", "c", "", "custom context")
	flags.StringVarP(&app.WorkingDirectory, "cwd", "", "", "current working directory")
	flags.StringVarP(&app.EOL, "eol", "", fmt.Sprintln(), "custom EOL char sequence")
	flags.StringArrayVarP(&app.EnvFiles, "env-file", "e", []string{}, "one or more env file to load")
	flags.StringArrayVarP(&app.Files, "file", "f", []string{}, "one or more files to use")
	flags.StringArrayVarP(&app.FilePatterns, "files", "", []string{}, "one or more files in form of patterns to use")
	flags.StringVarP(&app.HomeDirectory, "home", "", "", "user's home directory")
	flags.BoolVarP(&app.SkipDefaultEnvFiles, "skip-env-files", "", false, "do not load default .env files")
	flags.Int64VarP(&app.MaxTokens, "max-tokens", "", 0, "maximum number of tokens")
	flags.StringVarP(&app.Model, "model", "m", "", "default chat model")
	flags.StringVarP(&app.OutputFile, "output", "o", "", "write output to this file")
	flags.StringVarP(&app.SystemPrompt, "system", "s", "", "custom system prompt")
	flags.StringVarP(&app.SystemRole, "system-role", "", "", "custom name/id of the system role")
	flags.Float64VarP(&app.Temperature, "temperature", "t", -1, "custom temperature value")
	flags.StringVarP(&app.TerminalFormatter, "terminal-formatter", "", "", "custom terminal formatter")
	flags.StringVarP(&app.TerminalStyle, "terminal-style", "", "", "custom terminal style")
	flags.BoolVarP(&app.Verbose, "verbose", "", false, "verbose output")

	// Initialize commands
	commands.Init_analize_Command(app, rootCmd)
	commands.Init_chat_Command(app, rootCmd)
	commands.Init_init_Command(app, rootCmd)
	commands.Init_list_Command(app, rootCmd)
	commands.Init_prompt_Command(app, rootCmd)
	commands.Init_reset_Command(app, rootCmd)
	commands.Init_update_Command(app, rootCmd)

	app.Log = log.New(app, "", log.Ldate|log.Ltime)

	app.Run()
}
