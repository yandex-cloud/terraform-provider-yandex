package prompt

import (
	tea "github.com/charmbracelet/bubbletea"
)

type Prompt struct {
	quitting       bool
	model          PromptModel
	Message        string
	theme          Theme
	teaProgramOpts []tea.ProgramOption
}

// New returns a *Prompt using the default style.
func New(opts ...Option) *Prompt {
	p := &Prompt{
		quitting:       false,
		theme:          ThemeDefault,
		teaProgramOpts: make([]tea.ProgramOption, 0),
	}

	for _, opt := range opts {
		opt(p)
	}
	return p
}

// Ask set prompt message
func (p *Prompt) Ask(message string) *Prompt {
	p.Message = message
	return p
}

// SetModel sets the model used by the prompt. In most cases you won't need to
// use this.
func (p *Prompt) SetModel(pm PromptModel) *Prompt {
	p.model = pm
	return p
}

// Run runs the program using the given model, blocking until the user chooses
// or exits.
func (p *Prompt) Run(pm PromptModel, opts ...tea.ProgramOption) (PromptModel, error) {
	p.model = pm

	tm, err := tea.NewProgram(p, opts...).Run()
	if err != nil {
		return nil, err
	}

	m, ok := tm.(Prompt)
	if !ok {
		return nil, ErrModelConversion
	}

	if m.model.Error() != nil {
		return nil, m.model.Error()
	}

	return m.model, nil
}
