package write

import "github.com/charmbracelet/bubbles/key"

type KeyMap struct {
	Confirm key.Binding
	Quit    key.Binding
}

func (k KeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Confirm, k.Quit}
}

func (k KeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Confirm, k.Quit},
	}
}

var DefaultKeyMap = KeyMap{
	Confirm: key.NewBinding(
		key.WithKeys("ctrl+d"),
		key.WithHelp("ctrl+d", "confirm"),
	),
	Quit: key.NewBinding(
		key.WithKeys("esc", "ctrl+c"),
		key.WithHelp("esc", "quit"),
	),
}
