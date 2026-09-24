package choose

import (
	"fmt"
	"strings"

	"github.com/cqroot/prompt/constants"
)

type Theme func([]Choice, int) string

var ThemeDefault = func(choices []Choice, cursor int) string {
	s := strings.Builder{}
	s.WriteString("\n")

	maxLen := 0
	for _, choice := range choices {
		if maxLen < len([]rune(choice.Text)) {
			maxLen = len([]rune(choice.Text))
		}
	}

	for i := 0; i < len(choices); i++ {
		text := choices[i].Text
		note := choices[i].Note
		if note != "" {
			note = strings.Repeat(" ", maxLen-len([]rune(text))+2) + constants.DefaultNoteStyle.Render(note)
		}
		choice := text + note

		if cursor == i {
			s.WriteString(constants.DefaultSelectedItemStyle.Render("• " + choice))
		} else {
			s.WriteString(constants.DefaultItemStyle.Render(fmt.Sprintf("  " + choice)))
		}
		s.WriteString("\n")
	}

	return s.String()
}

var ThemeArrow = func(choices []Choice, cursor int) string {
	s := strings.Builder{}
	s.WriteString("\n")

	maxLen := 0
	for _, choice := range choices {
		if maxLen < len([]rune(choice.Text)) {
			maxLen = len([]rune(choice.Text))
		}
	}

	for i := 0; i < len(choices); i++ {
		text := choices[i].Text
		note := choices[i].Note
		if note != "" {
			note = strings.Repeat(" ", maxLen-len([]rune(text))+2) + constants.DefaultNoteStyle.Render(note)
		}
		choice := text + note

		if cursor == i {
			s.WriteString(constants.DefaultSelectedItemStyle.Render(("❯ " + choice)))
		} else {
			s.WriteString(constants.DefaultItemStyle.Render(fmt.Sprintf("  " + choice)))
		}
		s.WriteString("\n")
	}

	return s.String()
}

var ThemeLine = func(choices []Choice, cursor int) string {
	s := strings.Builder{}

	result := make([]string, len(choices))
	for index, choice := range choices {
		if index == cursor {
			result[index] = constants.DefaultSelectedItemStyle.Render(choice.Text)
		} else {
			result[index] = constants.DefaultItemStyle.Render(choice.Text)
		}
	}
	s.WriteString(strings.Join(result, " / "))
	s.WriteString("\n")

	return s.String()
}
