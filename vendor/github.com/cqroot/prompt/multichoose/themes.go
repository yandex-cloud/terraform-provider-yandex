package multichoose

import (
	"fmt"
	"strings"

	"github.com/cqroot/prompt/constants"
)

type IsSelected func(int) bool

type Theme func([]string, int, IsSelected) string

func ThemeDefault(choices []string, cursor int, isSelected IsSelected) string {
	s := strings.Builder{}
	s.WriteString("\n")

	for i := 0; i < len(choices); i++ {
		if cursor == i {
			if isSelected(i) {
				s.WriteString(constants.DefaultSelectedItemStyle.Render(fmt.Sprintf("[x] %s", choices[i])))
			} else {
				s.WriteString(constants.DefaultSelectedItemStyle.Render(fmt.Sprintf("[•] %s", choices[i])))
			}
		} else {
			if isSelected(i) {
				s.WriteString(constants.DefaultItemStyle.Render(fmt.Sprintf("[x] %s", choices[i])))
			} else {
				s.WriteString(constants.DefaultItemStyle.Render(fmt.Sprintf("[ ] %s", choices[i])))
			}
		}
		s.WriteString("\n")
	}

	return s.String()
}

func ThemeDot(choices []string, cursor int, isSelected IsSelected) string {
	s := strings.Builder{}
	s.WriteString("\n")

	for i := 0; i < len(choices); i++ {
		var text string
		if isSelected(i) {
			text = "◉ " + choices[i]
		} else {
			text = "○ " + choices[i]
		}
		if cursor == i {
			s.WriteString(constants.DefaultSelectedItemStyle.Render(text))
		} else {
			s.WriteString(constants.DefaultItemStyle.Render(text))
		}
		s.WriteString("\n")
	}

	return s.String()
}
