package write

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/cqroot/prompt/constants"
)

type Model struct {
	textarea     textarea.Model
	validateFunc ValidateFunc

	quitting       bool
	err            error
	keys           KeyMap
	showHelp       bool
	help           help.Model
	teaProgramOpts []tea.ProgramOption
}

func New(defaultValue string, opts ...Option) *Model {
	ta := textarea.New()
	ta.Placeholder = defaultValue
	ta.ShowLineNumbers = false
	ta.Focus()

	m := &Model{
		textarea:       ta,
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
	if m.textarea.Value() == "" {
		return m.textarea.Placeholder
	} else {
		return m.textarea.Value()
	}
}

func (m Model) DataString() string {
	data := m.Data()
	if strings.Contains(data, "\n") {
		return fmt.Sprintf("...(%d bytes)", len(m.Data()))
	} else {
		return data
	}
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

func (m Model) KeyBindings() []key.Binding {
	return nil
}

func (m *Model) WithValidateFunc(vf ValidateFunc) *Model {
	m.validateFunc = vf
	return m
}

func (m Model) Init() tea.Cmd {
	return textarea.Blink
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.help.Width = msg.Width

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Confirm):
			if m.err == nil && m.validateFunc != nil {
				currVal := m.textarea.Value()
				if currVal == "" {
					currVal = m.textarea.Placeholder
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
	}

	var cmd tea.Cmd

	m.textarea, cmd = m.textarea.Update(msg)
	return m, cmd
}

func (m Model) View() string {
	view := "\n" + m.textarea.View()

	if m.textarea.Value() != "" && m.validateFunc != nil {
		err := m.validateFunc(m.textarea.Value())
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
