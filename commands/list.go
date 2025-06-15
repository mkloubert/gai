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
	"sort"

	"github.com/joho/godotenv"
	"github.com/mkloubert/gai/types"
	"github.com/spf13/cobra"
)

func init_list_env_Command(app *types.AppContext, parentCmd *cobra.Command) {
	var noSort bool

	var listEnvCmd = &cobra.Command{
		Use:     "env",
		Aliases: []string{"a"},
		Short:   "List env vars",
		Long:    `Lists all environment variables for the app in an ordered list.`,
		Run: func(cmd *cobra.Command, args []string) {
			keys := make([]string, 0)
			for key, _ := range app.EnvVars {
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

	init_list_env_Command(app, listCmd)

	parentCmd.AddCommand(
		listCmd,
	)
}
