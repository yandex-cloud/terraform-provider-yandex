package prompt

import (
	tea "github.com/charmbracelet/bubbletea"
)

type PromptModel interface {
	tea.Model
	DataString() string // Returns a string for display in the result position.
	Quitting() bool
	Error() error
	TeaProgramOpts() []tea.ProgramOption
}

func (p Prompt) Init() tea.Cmd {
	return nil
}

func (p Prompt) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	model, cmd := p.model.Update(msg)
	p.model = model.(PromptModel)
	return p, cmd
}

func (p Prompt) View() string {
	if p.model.Error() != nil {
		return p.theme(p.Message, StateError, p.model.DataString())
	} else if p.model.Quitting() {
		return p.theme(p.Message, StateFinish, p.model.DataString())
	} else {
		return p.theme(p.Message, StateNormal, p.model.View())
	}
}
