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

package types

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	ignore "github.com/sabhiram/go-gitignore"
)

// GitClient handles git operations for an `AppContext`.
type GitClient struct {
	app AppContext
	dir string
}

// Dir returns the root directory of the underlying repository.
func (g *GitClient) Dir() string {
	return g.dir
}

// GetAllCommits returns all commits.
func (g *GitClient) GetAllCommits() ([]*GitCommit, error) {
	commits := make([]*GitCommit, 0)

	cmd := exec.Command("git", "log", "--pretty=format:%H")
	cmd.Dir = g.dir

	output, err := cmd.Output()
	if err != nil {
		return commits, err
	}

	commitsHashes := strings.SplitSeq(strings.TrimSpace(string(output)), "\n")
	for hash := range commitsHashes {
		hash = strings.TrimSpace(strings.ToLower(hash))
		if hash == "" {
			continue
		}

		commits = append(commits, &GitCommit{
			git:  g,
			hash: hash,
		})
	}

	return commits, nil
}

// GetChangedFiles returns the list of changed files that are not staged yet.
func (g *GitClient) GetChangedFiles() ([]*GitFile, error) {
	changedFiles := make([]*GitFile, 0)

	repoDir := g.dir

	cmd := exec.Command("git", "status", "--porcelain")
	cmd.Dir = repoDir

	var out bytes.Buffer
	cmd.Stdout = &out

	err := cmd.Run()
	if err != nil {
		return changedFiles, err
	}

	lines := strings.SplitSeq(out.String(), "\n")
	for line := range lines {
		if len(line) < 4 {
			continue
		}

		status := strings.ToUpper(line[:2])
		file := strings.TrimSpace(line[3:])

		// " M" = changed in working directory, "M " = changed in staging area
		// other stati: (A, D, etc.)
		if status == " M" || status == "MM" || status == "AM" || status == "A " || status == "D " {
			changedFiles = append(changedFiles, &GitFile{
				git:          g,
				name:         file,
				status:       "changed",
				changeStatus: status,
			})
		}
	}

	return changedFiles, nil
}

// GetFiles returns list of files related to this client / repository.
func (g *GitClient) GetFiles() ([]*GitFile, error) {
	app := g.app
	repoDir := g.dir

	files := make([]*GitFile, 0)

	appFiles, err := app.GetFiles()
	if err != nil {
		return files, err // could not get app files
	}

	gitignore, err := g.GetGitIgnore()
	if err != nil {
		return files, err // could not create gitignore instance
	}
	doIgnore := func(p string) bool {
		if gitignore == nil {
			return false
		}
		return gitignore.MatchesPath(p)
	}

	if len(appFiles) == 0 {
		// take all files of the repository

		err := filepath.WalkDir(repoDir, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}

			if !d.IsDir() {
				relPath, err := filepath.Rel(repoDir, path)
				if err != nil {
					return err
				}

				if !doIgnore(relPath) {
					appFiles = append(appFiles, path)
				}
			}

			return nil
		})

		if err != nil {
			return files, err
		}
	}

	// must be part of the repository
	requiredPrefix := fmt.Sprintf("%s%c", repoDir, os.PathSeparator)
	repoFiles := make([]string, 0)
	for _, f := range appFiles {
		if strings.HasPrefix(f, requiredPrefix) {
			repoFiles = append(repoFiles, f)
		}
	}

	for _, f := range repoFiles {
		if !doIgnore(f) {
			relPath, err := filepath.Rel(repoDir, f)
			if err != nil {
				return files, err
			}

			files = append(files, &GitFile{
				git:  g,
				name: relPath,
			})
		}
	}

	return files, nil
}

// GetLatestCommit tries to detect the latest commit.
func (g *GitClient) GetLatestCommit() (*GitCommit, error) {
	commit := &GitCommit{
		git: g,
	}

	headCmd := exec.Command("git", "rev-parse", "HEAD")
	headCmd.Dir = g.dir

	headOut, err := headCmd.Output()
	if err != nil {
		return commit, err
	}

	commit.hash = strings.TrimSpace(string(headOut))

	return commit, nil
}

// GetStagedFiles returns the list of staged files.
func (g *GitClient) GetStagedFiles() ([]*GitFile, error) {
	gitFiles := make([]*GitFile, 0)

	cmd := exec.Command("git", "diff", "--cached", "--name-status")
	cmd.Dir = g.dir

	output, err := cmd.Output()
	if err != nil {
		return gitFiles, err
	}

	files := strings.SplitSeq(strings.TrimSpace(string(output)), "\n")
	for file := range files {
		file = strings.TrimSpace(file)
		if file == "" {
			continue
		}

		var filename string
		var stageStatus string

		fields := strings.Fields(file)
		if len(fields) >= 2 {
			stageStatus = strings.TrimSpace(strings.ToUpper(fields[0]))
			filename = strings.TrimSpace(fields[1])
		} else {
			filename = file
		}

		gitFiles = append(gitFiles, &GitFile{
			git:         g,
			name:        filename,
			stageStatus: stageStatus,
			status:      "staged",
		})
	}

	return gitFiles, nil
}

// GetGitIgnore loads .gitignore file if available.
func (g *GitClient) GetGitIgnore() (*ignore.GitIgnore, error) {
	dir := g.dir
	var gitignore *ignore.GitIgnore

	file := filepath.Join(dir, ".gitignore")

	if _, err := os.Stat(dir); os.IsNotExist(err) {
		// file not found
	} else if err != nil {
		// other error
		return gitignore, err
	}

	rawContent, err := os.ReadFile(file)
	if err != nil {
		return gitignore, err // could not load content
	}

	lines := make([]string, 0)
	for _, l := range strings.Split(string(rawContent), "\n") {
		l = strings.TrimSpace(l)
		if l != "" {
			lines = append(lines, l)
		}
	}

	gitignore = ignore.CompileIgnoreLines(lines...)

	return gitignore, nil
}
