package prompt

import (
	"strings"

	"github.com/cqroot/prompt/constants"
)

type State int

const (
	StateNormal State = iota
	StateFinish
	StateError
)

type Theme func(string, State, string) string

func ThemeDefault(msg string, state State, model string) string {
	s := strings.Builder{}

	switch state {
	case StateNormal:
		s.WriteString(constants.DefaultFinishPromptPrefixStyle.Render("?"))
	case StateFinish:
		s.WriteString(constants.DefaultFinishPromptPrefixStyle.Render("✔"))
	case StateError:
		s.WriteString(constants.DefaultErrorPromptPrefixStyle.Render("✖"))
	}

	s.WriteString(" ")
	s.WriteString(msg)
	s.WriteString(" ")

	if state == StateNormal {
		s.WriteString(constants.DefaultNormalPromptSuffixStyle.Render("›"))
		s.WriteString(" ")
		s.WriteString(model)
	} else {
		s.WriteString(constants.DefaultFinishPromptSuffixStyle.Render("…"))
		s.WriteString(" ")
		s.WriteString(model)
		s.WriteString("\n")
	}

	return s.String()
}

// ThemeDefaultClear is basically the same as ThemeDefault, but it will
// clear the screen after the selection is completed or after exiting.
func ThemeDefaultClear(msg string, state State, model string) string {
	if state == StateFinish || state == StateError {
		return ""
	}
	return ThemeDefault(msg, state, model)
}
