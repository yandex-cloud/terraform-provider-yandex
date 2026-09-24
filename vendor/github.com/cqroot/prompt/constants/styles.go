package constants

import (
	"github.com/charmbracelet/lipgloss"
)

var (
	DefaultNormalPromptPrefixStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("14"))
	DefaultFinishPromptPrefixStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
	DefaultErrorPromptPrefixStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("1"))
	DefaultNormalPromptSuffixStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("6"))
	DefaultFinishPromptSuffixStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("6"))

	DefaultNoteStyle = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{
		Light: "#909090",
		Dark:  "#626262",
	})

	// for choose, multichoose
	DefaultItemStyle         = lipgloss.NewStyle()
	DefaultSelectedItemStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("14"))
	DefaultChoiceStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("14"))
)
