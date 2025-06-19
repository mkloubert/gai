package types

import (
	"github.com/alecthomas/chroma/v2/quick"
)

// ChromaSettings stores settings for syntax highlighted console output.
type ChromaSettings struct {
	App *AppContext
	// Formatter stores the name of the chroma formatter.
	Formatter string
	// Style stores the name of the chroma style.
	Style string
}

// Highlight outputs a string highlighted in the defined language.
func (cs *ChromaSettings) Highlight(s string, language string) {
	err := quick.Highlight(cs.App, s, language, cs.Formatter, cs.Style)
	if err != nil {
		cs.App.Write([]byte(s))
	}
}

// HighlightMarkdown outputs a string highlighted in Markdown.
func (cs *ChromaSettings) HighlightMarkdown(s string) {
	cs.Highlight(s, "markdown")
}
