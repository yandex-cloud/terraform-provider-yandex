package input

import (
	"strings"
	"unicode"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/cqroot/prompt/constants"
)

type Model struct {
	df           string
	textInput    textinput.Model
	validateFunc ValidateFunc
	inputMode    InputMode

	quitting       bool
	err            error
	keys           KeyMap
	showHelp       bool
	help           help.Model
	teaProgramOpts []tea.ProgramOption
}

func New(defaultValue string, opts ...Option) *Model {
	ti := textinput.New()
	ti.Placeholder = defaultValue
	ti.Focus()
	ti.CharLimit = 156
	ti.Width = 40
	ti.Prompt = ""

	m := &Model{
		textInput:      ti,
		df:             defaultValue,
		inputMode:      InputAll,
		quitting:       false,
		err:            nil,
		keys:           DefaultKeyMap,
		showHelp:       false,
		help:           help.New(),
		teaProgramOpts: make([]tea.ProgramOption, 0),
	}

	for _, opt := range opts {
		opt(m)
	}

	return m
}

func (m Model) Data() string {
	if m.textInput.Value() == "" {
		return m.textInput.Placeholder
	} else {
		return m.textInput.Value()
	}
}

func (m Model) DataString() string {
	if m.textInput.EchoMode == EchoNormal {
		return m.Data()
	}
	m.textInput.Blur()
	str := m.textInput.View()
	m.textInput.Focus()
	return str
}

func (m Model) Quitting() bool {
	return m.quitting
}

func (m Model) Error() error {
	return m.err
}

func (m Model) TeaProgramOpts() []tea.ProgramOption {
	return m.teaProgramOpts
}

func (m *Model) WithInputMode(mode InputMode) *Model {
	m.inputMode = mode
	return m
}

func (m *Model) WithEchoMode(mode EchoMode) *Model {
	m.textInput.EchoMode = mode
	return m
}

func (m *Model) WithValidateFunc(vf ValidateFunc) *Model {
	m.validateFunc = vf
	return m
}

func (m Model) Init() tea.Cmd {
	return textinput.Blink
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.help.Width = msg.Width

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Confirm):
			if m.err == nil && m.validateFunc != nil {
				currVal := m.textInput.Value()
				if currVal == "" {
					currVal = m.textInput.Placeholder
				}
				m.err = m.validateFunc(currVal)
			}
			m.quitting = true
			return m, tea.Quit

		case key.Matches(msg, m.keys.Quit):
			m.quitting = true
			m.err = constants.ErrUserQuit
			return m, tea.Quit
		}

		if m.inputMode == InputNumber || m.inputMode == InputInteger {
			keypress := msg.String()
			if len(keypress) == 1 {
				if keypress == "." {
					if m.inputMode != InputNumber ||
						strings.Contains(m.textInput.Value(), ".") {
						return m, nil
					}
				} else {
					if !unicode.IsNumber([]rune(keypress)[0]) {
						return m, nil
					}
				}
			}
		}
	}

	var cmd tea.Cmd
	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

func (m Model) View() string {
	view := m.textInput.View()

	if m.textInput.Value() != "" && m.validateFunc != nil {
		err := m.validateFunc(m.textInput.Value())
		if err != nil {
			view = view + constants.DefaultErrorPromptPrefixStyle.Render("\n✖  ") +
				constants.DefaultNoteStyle.Render(err.Error())
		}
	}

	if m.showHelp {
		view += "\n\n"
		view += m.help.View(m.keys)
	}

	return view
}
