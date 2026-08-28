package main

import (
	"io"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// theme is the shared Charm palette for plain status lines (ensure, check,
// commit confirmations) and for the file preview inside the interactive
// editor. It is scoped to a specific writer so color auto-disables when
// output isn't a terminal — piped into a script, captured by `just switch`,
// or a bytes.Buffer in tests — exactly like huh's own theme does for forms.
type theme struct {
	renderer *lipgloss.Renderer
	ok       lipgloss.Style
	warn     lipgloss.Style
	fail     lipgloss.Style
	info     lipgloss.Style
	muted    lipgloss.Style
	section  lipgloss.Style
}

func newTheme(w io.Writer) *theme {
	r := lipgloss.NewRenderer(w)
	return &theme{
		renderer: r,
		ok:       r.NewStyle().Foreground(lipgloss.Color("42")).Bold(true),
		warn:     r.NewStyle().Foreground(lipgloss.Color("214")).Bold(true),
		fail:     r.NewStyle().Foreground(lipgloss.Color("204")).Bold(true),
		info:     r.NewStyle().Foreground(lipgloss.Color("111")),
		muted:    r.NewStyle().Foreground(lipgloss.Color("243")),
		section:  r.NewStyle().Foreground(lipgloss.Color("111")).Bold(true),
	}
}

// renderPreview renders the whole file as it currently stands in memory —
// comments dimmed, section headers accented — so the editor always shows the
// complete document, not just the section being touched.
func (t *theme) renderPreview(cfg Config) string {
	data, err := encodeConfig(cfg)
	if err != nil {
		return t.fail.Render("encoding error: " + err.Error())
	}
	lines := strings.Split(strings.TrimRight(string(data), "\n"), "\n")
	for i, line := range lines {
		switch {
		case len(line) > 0 && line[0] == '#':
			lines[i] = t.muted.Render(line)
		case len(line) > 0 && line[0] == '[':
			lines[i] = t.section.Render(line)
		}
	}
	return strings.Join(lines, "\n")
}
