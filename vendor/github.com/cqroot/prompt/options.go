package prompt

import tea "github.com/charmbracelet/bubbletea"

type Option func(*Prompt)

func WithTheme(theme Theme) Option {
	return func(p *Prompt) {
		p.theme = theme
	}
}

func WithTeaProgramOpts(opts ...tea.ProgramOption) Option {
	return func(p *Prompt) {
		p.teaProgramOpts = append(p.teaProgramOpts, opts...)
	}
}
