package choose

import "github.com/charmbracelet/bubbles/key"

type KeyMap struct {
	Prev    key.Binding
	Next    key.Binding
	Help    key.Binding
	Confirm key.Binding
	Quit    key.Binding
}

func (k KeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Help, k.Confirm, k.Quit}
}

func (k KeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Prev, k.Next},            // first column
		{k.Help, k.Confirm, k.Quit}, // second column
	}
}

var (
	DefaultKeyMap = KeyMap{
		Help: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "toggle help"),
		),
		Confirm: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "confirm"),
		),
		Quit: key.NewBinding(
			key.WithKeys("q", "esc", "ctrl+c"),
			key.WithHelp("q", "quit"),
		),
		Prev: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("↑/k", "move up"),
		),
		Next: key.NewBinding(
			key.WithKeys("down", "j", "tab"),
			key.WithHelp("↓/j/tab", "move down"),
		),
	}

	HorizontalKeyMap = KeyMap{
		Help: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "toggle help"),
		),
		Confirm: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "confirm"),
		),
		Quit: key.NewBinding(
			key.WithKeys("q", "esc", "ctrl+c"),
			key.WithHelp("q", "quit"),
		),
		Prev: key.NewBinding(
			key.WithKeys("left", "h"),
			key.WithHelp("←/h", "move left"),
		),
		Next: key.NewBinding(
			key.WithKeys("right", "l", "tab", " "),
			key.WithHelp("→/l/tab/space", "move right"),
		),
	}
)
