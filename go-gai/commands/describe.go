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
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/mkloubert/gai/types"
	"github.com/spf13/cobra"
)

type imageDescriptionResponse struct {
	FileModifiationTime string                                   `json:"file_modifiation_time,omitempty"`
	Filename            string                                   `json:"filename,omitempty"`
	Filesize            int64                                    `json:"filesize,omitempty"`
	ImageInformation    imageDescriptionResponseImageInformation `json:"image_information,omitempty"`
}

type imageDescriptionResponseImageInformation struct {
	DetailedDescription string   `json:"detailed_description"`
	Tags                []string `json:"tags"`
	Title               string   `json:"title"`
}

func init_describe_images_Command(app *types.AppContext, parentCmd *cobra.Command) {
	var forceUpdate bool
	var maxTags uint16
	var minTags uint16
	var updateExisting bool

	var initCodeCmd = &cobra.Command{
		Use:     "images",
		Aliases: []string{"images", "image", "img", "imgs", "i"},
		Short:   "Describe image",
		Long:    `Describes images with tags.`,
		Run: func(cmd *cobra.Command, args []string) {
			app.InitAI()

			files, err := app.GetFiles()
			app.CheckIfError(err)

			db, err := app.OpenSQLDatabase()
			app.CheckIfError(err)

			if db != nil {
				createTable := `CREATE TABLE IF NOT EXISTS images (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  file_path TEXT NOT NULL,
  last_filesize INTEGER NOT NULL,
  last_modified DATETIME NOT NULL,
  title TEXT NOT NULL,
  description TEXT NOT NULL,
  tags TEXT NOT NULL,
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP NOT NULL,
  updated_at DATETIME
);`
				_, err = db.Exec(createTable)
				app.CheckIfError(err)

				createIndex := `CREATE UNIQUE INDEX IF NOT EXISTS idx_images_file_path ON images(file_path);`
				_, err = db.Exec(createIndex)
				app.CheckIfError(err)

				defer func() {
					db.Close()
				}()
			}

			if len(files) == 0 {
				app.CheckIfError(errors.New("no files found or defined"))
			}

			responseSchema, responseSchemaName, err := app.GetResponseSchema()
			app.CheckIfError(err)

			outputLanguage := strings.TrimSpace(app.OutputLanguage)

			lang := "english"
			if outputLanguage != "" {
				lang = outputLanguage
			}

			prompt, err := app.GetInput(args)
			app.CheckIfError(err)

			prompt = strings.TrimSpace(prompt)
			if prompt == "" {
				prompt = "What is in this image?"
			}

			systemPrompt := fmt.Sprintf(`You are an AI assistant that helps users organize their photo collections.
For each provided image file, generate:
- A concise and informative description of the image in natural '%s' language, suitable for someone who cannot see the photo.
- A short and descriptive title of the main objects in the image.
- A set of relevant tags that summarize the main objects, themes, activities, and visual elements present in the image. The tags should be lowercase, and without special characters.
Be objective and accurate. Do not include personal opinions or assumptions that cannot be verified from the image itself.`, lang)

			if responseSchema == nil {
				// we want structured output

				responseSchema = &map[string]any{
					"type":     "object",
					"required": []string{"image_information", "tags", "title"},
					"properties": map[string]any{
						"image_information": map[string]any{
							"type":        "object",
							"description": "Information about the image.",
							"required":    []string{"detailed_description", "tags", "title"},
							"properties": map[string]any{
								"detailed_description": map[string]any{
									"description": "A detailed description what is in the image.",
									"type":        "string",
								},
								"tags": map[string]any{
									"type":    "array",
									"minimum": minTags,
									"maximum": maxTags,
									"items": map[string]any{
										"type":        "string",
										"description": "Word or small text that categorized the image.",
									},
								},
								"title": map[string]any{
									"description": "A short and descriptive title for the image.",
									"type":        "string",
								},
							},
						},
					},
				}
			}
			if strings.TrimSpace(responseSchemaName) == "" {
				responseSchemaName = "DescribeImageSchema"
			}

			outputError := func(err error) {
				errorObj := &map[string]any{
					"error": map[string]any{
						"message": err.Error(),
					},
				}

				data, err2 := json.Marshal(&errorObj)
				if err2 != nil {
					app.Writeln(err2)
				} else {
					app.Writeln(fmt.Sprintf("ERROR: %s", data))
				}
			}

			for _, f := range files {
				func() {
					info, err := os.Stat(f)
					if err != nil {
						outputError(err)
						return
					}

					// get file size and last update time
					filesize := info.Size()
					fileModTime := info.ModTime().UTC().Format(time.RFC3339)

					file, err := os.Open(f)
					if err != nil {
						outputError(err)
						return
					}

					defer file.Close()

					filename, err := filepath.Rel(app.WorkingDirectory, f)
					if err != nil {
						filename = f
					}

					if db != nil && !forceUpdate {
						// check for existing entries and if they should be updated

						var lastFilesize int64
						var lastModified string

						err := db.QueryRow(
							`SELECT last_filesize, last_modified FROM images
WHERE file_path = ?;`,
							filename,
						).Scan(&lastFilesize, &lastModified)

						if err == nil {
							// exists
							if !updateExisting {
								return // ... but do not update
							}
						} else if err != sql.ErrNoRows {
							app.CheckIfError(err)
						}
					}

					promptOptions := make([]types.AIClientPromptOptions, 0)
					promptOptions = append(promptOptions, types.AIClientPromptOptions{
						Files:              &[]io.Reader{file},
						ResponseSchema:     responseSchema,
						ResponseSchemaName: &responseSchemaName,
						SystemPrompt:       &systemPrompt,
					})

					response, err := app.AI.Prompt(prompt, promptOptions...)
					if err != nil {
						outputError(err)
						return
					}

					// ensure we have correct response ...
					var imageDescription imageDescriptionResponse
					err = json.Unmarshal([]byte(response.Content), &imageDescription)
					if err != nil {
						outputError(err)
						return
					}

					imageDescription.Filename = filename
					imageDescription.Filesize = filesize
					imageDescription.FileModifiationTime = fileModTime

					// ... and finally a cleaned JSON
					cleanJson, err := json.Marshal(&imageDescription)
					if err != nil {
						outputError(err)
						return
					}

					app.Writeln(string(cleanJson))

					if db != nil {
						func() {
							stmt, err := db.Prepare(`INSERT INTO images
(file_path, title, description, tags, last_filesize, last_modified) VALUES (?, ?, ?, ?, ?, ?)
ON CONFLICT(file_path) DO UPDATE SET
    description=excluded.description,
    tags=excluded.tags,
	title=excluded.title,
    last_filesize=excluded.last_filesize,
    last_modified=excluded.last_modified,
	updated_at=CURRENT_TIMESTAMP;`)
							app.CheckIfError(err)

							defer stmt.Close()

							_, err = stmt.Exec(
								imageDescription.Filename,
								imageDescription.ImageInformation.Title,
								imageDescription.ImageInformation.DetailedDescription,
								strings.Join(imageDescription.ImageInformation.Tags, ","),
								filesize,
								fileModTime,
							)
							app.CheckIfError(err)
						}()
					}
				}()
			}
		},
	}

	initCodeCmd.Flags().BoolVarP(&forceUpdate, "force-update", "", false, "")
	initCodeCmd.Flags().Uint16VarP(&maxTags, "max-tags", "", 10, "")
	initCodeCmd.Flags().Uint16VarP(&minTags, "min-tags", "", 1, "")
	initCodeCmd.Flags().BoolVarP(&updateExisting, "update-existing", "", false, "")

	app.WithDatabaseFlags(initCodeCmd)
	app.WithLanguageFlags(initCodeCmd)

	parentCmd.AddCommand(
		initCodeCmd,
	)
}

// Init_describe_Command initializes the `describe` command.
func Init_describe_Command(app *types.AppContext, parentCmd *cobra.Command) {
	var initCmd = &cobra.Command{
		Use:     "describes [resource]",
		Aliases: []string{"describes", "desc", "d"},
		Short:   "Describe",
		Long:    `Describes a resource.`,
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}

	init_describe_images_Command(app, initCmd)

	parentCmd.AddCommand(
		initCmd,
	)
}
