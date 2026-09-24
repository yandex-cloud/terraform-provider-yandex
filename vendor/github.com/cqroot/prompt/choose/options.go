package choose

import tea "github.com/charmbracelet/bubbletea"

type Option func(*Model)

func WithTeaProgramOpts(opts ...tea.ProgramOption) Option {
	return func(m *Model) {
		m.teaProgramOpts = append(m.teaProgramOpts, opts...)
	}
}

func WithHelp(show bool) Option {
	return func(m *Model) {
		m.showHelp = show
	}
}

func WithTheme(theme Theme) Option {
	return func(m *Model) {
		m.theme = theme
	}
}

func WithKeyMap(keyMap KeyMap) Option {
	return func(m *Model) {
		m.keys = keyMap
	}
}

func WithDefaultIndex(index int) Option {
	return func(m *Model) {
		m.cursor = index
	}
}
