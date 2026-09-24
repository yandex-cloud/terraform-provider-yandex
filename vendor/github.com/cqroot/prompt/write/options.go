package write

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

func WithKeyMap(keyMap KeyMap) Option {
	return func(m *Model) {
		m.keys = keyMap
	}
}

// Default is 400.
// https://github.com/charmbracelet/bubbles/blob/master/textarea/textarea.go#L23
func WithCharLimit(limit int) Option {
	return func(m *Model) {
		m.textarea.CharLimit = limit
	}
}

func WithLineNumbers(enable bool) Option {
	return func(m *Model) {
		m.textarea.ShowLineNumbers = true
	}
}

type ValidateFunc func(string) error

func WithValidateFunc(vf ValidateFunc) Option {
	return func(m *Model) {
		m.WithValidateFunc(vf)
	}
}

func WithWidth(width int) Option {
	return func(m *Model) {
		m.textarea.SetWidth(width)
	}
}
