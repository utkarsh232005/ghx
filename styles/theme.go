package styles

import (
	"github.com/charmbracelet/lipgloss"
)

type Theme struct {
	Primary   lipgloss.Style
	Secondary lipgloss.Style
	Accent    lipgloss.Style
	Success   lipgloss.Style
	Warning   lipgloss.Style
	Error     lipgloss.Style
	Muted     lipgloss.Style

	Border     lipgloss.Style
	Title      lipgloss.Style
	Header     lipgloss.Style
	Footer     lipgloss.Style
	MenuItem   lipgloss.Style
	Selected   lipgloss.Style

	Text lipgloss.Style
	Bold lipgloss.Style

	Added     lipgloss.Style
	Modified  lipgloss.Style
	Deleted   lipgloss.Style
	Untracked lipgloss.Style
	Staged    lipgloss.Style

	Box   lipgloss.Style
	List  lipgloss.Style
	Input lipgloss.Style
	Help  lipgloss.Style
}

func NewTheme() *Theme {
	base := lipgloss.NewStyle()

	return &Theme{
		Primary:   base.Foreground(lipgloss.Color("#7C3AED")).Bold(true),
		Secondary: base.Foreground(lipgloss.Color("#64748B")),
		Accent:    base.Foreground(lipgloss.Color("#06B6D4")),
		Success:   base.Foreground(lipgloss.Color("#22C55E")),
		Warning:   base.Foreground(lipgloss.Color("#F59E0B")),
		Error:     base.Foreground(lipgloss.Color("#EF4444")),
		Muted:     base.Foreground(lipgloss.Color("#6B7280")),

		Border: base.Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#374151")).
			Padding(0, 1),
		Title: base.Foreground(lipgloss.Color("#F8FAFC")).
			Background(lipgloss.Color("#1E293B")).
			Padding(0, 2).
			Bold(true),
		Header:   base.Foreground(lipgloss.Color("#94A3B8")).Padding(0, 1),
		Footer:   base.Foreground(lipgloss.Color("#64748B")).Background(lipgloss.Color("#1E293B")).Padding(0, 1),
		MenuItem: base.Padding(0, 2),
		Selected: base.Foreground(lipgloss.Color("#22C55E")).
			Background(lipgloss.Color("#1E293B")).
			Padding(0, 2).
			Bold(true),

		Text: base.Foreground(lipgloss.Color("#F8FAFC")),
		Bold: base.Bold(true),

		Added:     base.Foreground(lipgloss.Color("#22C55E")),
		Modified:  base.Foreground(lipgloss.Color("#F59E0B")),
		Deleted:   base.Foreground(lipgloss.Color("#EF4444")),
		Untracked: base.Foreground(lipgloss.Color("#6B7280")),
		Staged:    base.Foreground(lipgloss.Color("#22C55E")),

		Box: base.Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#374151")).
			Padding(1, 2),
		List:  base.Padding(0, 1),
		Input: base.Border(lipgloss.NormalBorder()).BorderForeground(lipgloss.Color("#374151")).Padding(0, 1),
		Help:  base.Foreground(lipgloss.Color("#64748B")).Padding(0, 1),
	}
}

func (t *Theme) StatusIcon(status string) string {
	switch status {
	case "A":
		return t.Added.Render("A")
	case "M":
		return t.Modified.Render("M")
	case "D":
		return t.Deleted.Render("D")
	case "?":
		return t.Untracked.Render("?")
	default:
		return t.Muted.Render(status)
	}
}

func (t *Theme) Checkbox(checked bool) string {
	if checked {
		return t.Success.Render("[x]")
	}
	return t.Muted.Render("[ ]")
}

func (t *Theme) MenuItemWithSelector(item string, selected bool) string {
	if selected {
		return t.Selected.Render("> " + item)
	}
	return t.MenuItem.Render("  " + item)
}
