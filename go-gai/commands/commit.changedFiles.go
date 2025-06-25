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

	tea "github.com/charmbracelet/bubbletea"
	"github.com/mkloubert/gai/types"
)

type stageChangedFilesModelItem struct {
	checked bool
	file    *types.GitFile
	label   string
}

type stageChangedFilesModel struct {
	cursor int
	done   bool
	items  []stageChangedFilesModelItem
}

func (m stageChangedFilesModel) Init() tea.Cmd {
	return nil
}

func (m stageChangedFilesModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.done {
		return m, tea.Quit
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.items)-1 {
				m.cursor++
			}
		case " ":
			m.items[m.cursor].checked = !m.items[m.cursor].checked
		case "enter":
			m.done = true
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m stageChangedFilesModel) View() string {
	if m.done {
		s := "Files to stage:\n"
		for _, it := range m.items {
			if it.checked {
				s += fmt.Sprintf(" [x] %s\n", it.label)
			}
		}
		return s
	}

	s := "Do want want to stage following files instead? (↑/↓ to navigate, [Space] to select, [Enter] to approve)\n\n"
	for i, it := range m.items {
		cursor := "  "
		if m.cursor == i {
			cursor = "> "
		}

		checked := "[ ]"
		if it.checked {
			checked = "[x]"
		}

		s += fmt.Sprintf("%s%s %s\n", cursor, checked, it.label)
	}

	s += "\n[q] to exit\n"
	return s
}
