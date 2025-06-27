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
	"strings"
)

// GitCommit handles a file in a git repository.
type GitCommit struct {
	git  *GitClient
	hash string
}

// Diff returns the diff of a file compared with this commit.
func (gc *GitCommit) Diff(f *GitFile) (string, error) {
	return f.CompareWith(gc)
}

// GetFiles returns the list of files of this commit.
func (gc *GitCommit) GetFiles() ([]*GitFile, error) {
	git := gc.git

	cmd := git.CreateExecCommand("git", "ls-tree", "--name-only", "-r", gc.hash)

	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	gitFiles := make([]*GitFile, 0)

	files := strings.SplitSeq(strings.TrimSpace(string(output)), "\n")
	for f := range files {
		f = strings.TrimSpace(f)
		if f == "" {
			continue
		}

		gitFiles = append(gitFiles, &GitFile{
			commit: gc,
			git:    git,
			name:   f,
			status: "commited",
		})
	}

	return gitFiles, nil
}

// Hash returns the current hash of the commit.
func (gc *GitCommit) Hash() string {
	return gc.hash
}
