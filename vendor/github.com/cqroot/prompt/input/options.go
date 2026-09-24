package input

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

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

type EchoMode = textinput.EchoMode

const (
	// EchoNormal displays text as is. This is the default behavior.
	EchoNormal EchoMode = textinput.EchoNormal

	// EchoPassword displays the EchoCharacter mask instead of actual
	// characters.  This is commonly used for password fields.
	EchoPassword EchoMode = textinput.EchoPassword

	// EchoNone displays nothing as characters are entered. This is commonly
	// seen for password fields on the command line.
	EchoNone EchoMode = textinput.EchoNone
)

func WithEchoMode(mode EchoMode) Option {
	return func(m *Model) {
		m.WithEchoMode(mode)
	}
}

type InputMode int

const (
	InputAll     InputMode = iota // allow any input.
	InputInteger                  // only integers can be entered.
	InputNumber                   // only integers and decimals can be entered.
)

func WithInputMode(mode InputMode) Option {
	return func(m *Model) {
		m.WithInputMode(mode)
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
		m.textInput.Width = width
	}
}
