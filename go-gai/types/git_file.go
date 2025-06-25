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
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// GitFile handles a file in a git repository.
type GitFile struct {
	changeStatus string
	commit       *GitCommit
	git          *GitClient
	name         string
	stageStatus  string
	status       string
}

// ChangeStatus returns the status if changed.
func (gf *GitFile) ChangeStatus() string {
	return gf.changeStatus
}

// Commit returns the full path of the file.
func (gf *GitFile) Commit() *GitCommit {
	return gf.commit
}

// CompareWith returns the diff of the file based on a specific commit.
func (gf *GitFile) CompareWith(c *GitCommit) (string, error) {
	cmd := exec.Command("git", "diff", "--cached", c.hash, "--", gf.name)

	var out bytes.Buffer
	cmd.Stdout = &out

	err := cmd.Run()

	return out.String(), err
}

// FullName returns the full path of the file.
func (gf *GitFile) FullName() string {
	g := gf.git

	return filepath.Join(g.Dir(), gf.name)
}

// GetContent returns the content of the file based on its status.
func (gf *GitFile) GetContent() ([]byte, error) {
	git := gf.git

	if gf.IsStaged() {
		cmd := exec.Command("git", "show", ":"+gf.name)
		cmd.Dir = git.dir

		out := &bytes.Buffer{}
		cmd.Stdout = out

		err := cmd.Run()

		return out.Bytes(), err
	}

	return gf.GetCurrentContent()
}

// GetCurrentContent returns the current content of the file.
func (gf *GitFile) GetCurrentContent() ([]byte, error) {
	return os.ReadFile(gf.FullName())
}

// GetLatestContent returns the latest content of the file in the history.
func (gf *GitFile) GetLatestContent() ([]byte, error) {
	git := gf.git

	latestCommit, err := git.GetLatestCommit()
	if err != nil {
		return nil, err
	}

	cmd := exec.Command("git", "show", fmt.Sprintf("%s:%s", latestCommit.hash, gf.name))
	cmd.Dir = git.dir

	var out bytes.Buffer
	cmd.Stdout = &out

	err = cmd.Run()
	if err != nil {
		noExistingMessage := fmt.Sprintf(
			"fatal: path '%s' does not exist in '%s'",
			gf.name,
			latestCommit.hash,
		)

		str := strings.TrimSpace(out.String())
		if str == noExistingMessage {
			return nil, nil
		}

		return nil, err
	}

	return out.Bytes(), nil
}

// IsStaged returns returns `true` if this file is staged.
func (gf *GitFile) IsStaged() bool {
	return gf.status == "staged"
}

// Name returns the name of the file.
func (gf *GitFile) Name() string {
	return gf.name
}

func (gf *GitFile) Refresh() error {
	if gf.IsStaged() {
		cmd := exec.Command("git", "diff", "--cached", "--name-status", gf.name)

		var out bytes.Buffer
		cmd.Stdout = &out

		err := cmd.Run()
		if err != nil {
			return err
		}

		gf.stageStatus = ""

		output := strings.TrimSpace(out.String())
		if output == "" {
			return nil // not staged
		}

		fields := strings.Fields(output)
		if len(fields) >= 2 {
			gf.stageStatus = strings.ToUpper(fields[0])
		}
	}

	return nil
}

func (gf *GitFile) Stage() error {
	cmd := exec.Command("git", "add", gf.name)
	_, err := cmd.CombinedOutput()

	if err != nil {
		return err
	}

	gf.status = "staged"

	return gf.Refresh()
}

// StageStatus returns the status if staged.
func (gf *GitFile) StageStatus() string {
	return gf.stageStatus
}

// Status returns the status.
func (gf *GitFile) Status() string {
	return gf.status
}
