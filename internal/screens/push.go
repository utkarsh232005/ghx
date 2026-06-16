package screens

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/KDM-cli/ghx/internal/db"
	"github.com/KDM-cli/ghx/internal/git"
	"github.com/KDM-cli/ghx/styles"
)

type pushState int

const (
	pushPreview pushState = iota
	pushConfirm
	pushing
	pushDone
)

type PushModel struct {
	theme       *styles.Theme
	db          *db.DB
	state       pushState
	status      *git.Status
	commits     []git.CommitInfo
	selected    int
	remote      string
	forcePush   bool
	setUpstream bool
	width       int
	height      int
	err         error
}

func NewPushModel(theme *styles.Theme, database *db.DB) PushModel {
	return PushModel{
		theme:  theme,
		db:     database,
		state:  pushPreview,
		remote: "origin",
	}
}

func (m PushModel) Init() tea.Cmd {
	return m.loadStatus
}

func (m PushModel) loadStatus() tea.Msg {
	status, err := git.GetStatus()
	if err != nil {
		return pushStatusMsg{err: err}
	}

	commits, err := git.GetLog(10)
	if err != nil {
		return pushStatusMsg{err: err}
	}

	return pushStatusMsg{status: status, commits: commits}
}

type pushStatusMsg struct {
	status  *git.Status
	commits []git.CommitInfo
	err     error
}

type pushResultMsg struct {
	success bool
	err     error
}

func (m PushModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case pushStatusMsg:
		m.state = pushPreview
		if msg.err != nil {
			m.err = msg.err
			return m, nil
		}
		m.status = msg.status
		m.commits = msg.commits
		m.err = nil

	case pushResultMsg:
		m.state = pushDone
		m.err = msg.err
		if msg.success {
			return m, m.loadStatus
		}

	case tea.KeyMsg:
		// Reset error on any key
		m.err = nil

		switch msg.String() {
		case "enter":
			switch m.state {
			case pushPreview:
				if m.status != nil && m.status.Info.Ahead > 0 {
					m.state = pushConfirm
				}
			case pushConfirm:
				m.state = pushing
				return m, m.doPush
			case pushDone:
				m.state = pushPreview
				return m, m.loadStatus
			}

		case "tab":
			if m.state == pushConfirm {
				m.forcePush = !m.forcePush
			}

		case "u":
			if m.state == pushConfirm {
				m.setUpstream = !m.setUpstream
			}

		case "p":
			if m.state == pushPreview {
				return m, m.loadStatus
			}

		case "r":
			if m.state == pushDone {
				m.state = pushPreview
				return m, m.loadStatus
			}
		case "b":
			if m.state == pushPreview || m.state == pushDone {
				return m, func() tea.Msg {
					return Navigate(ScreenHome)
				}
			}
		}
	}

	return m, nil
}

func (m PushModel) doPush() tea.Msg {
	branch := ""
	if m.status != nil {
		branch = m.status.Info.Branch
	}

	if m.forcePush {
		// Force push
		if err := git.RunCommand("push", "--force", m.remote, branch); err != nil {
			return pushResultMsg{err: err}
		}
	} else {
		if err := git.Push(m.remote, branch); err != nil {
			return pushResultMsg{err: err}
		}
	}

	return pushResultMsg{success: true}
}

func (m PushModel) View() string {
	var b strings.Builder

	b.WriteString(m.theme.Title.Render("Push"))
	b.WriteString("\n\n")

	if m.err != nil {
		b.WriteString(m.theme.Error.Render("Error: " + m.err.Error()))
		b.WriteString("\n\n")
		b.WriteString(m.theme.Help.Render("r Retry   b Back"))
		return b.String()
	}

	switch m.state {
	case pushPreview:
		if m.status == nil {
			b.WriteString(m.theme.Muted.Render("Loading..."))
			return b.String()
		}

		// Branch info
		b.WriteString(m.theme.Bold.Render("Branch: "))
		b.WriteString(m.theme.Text.Render(m.status.Info.Branch))
		b.WriteString("\n")

		b.WriteString(m.theme.Bold.Render("Remote: "))
		b.WriteString(m.theme.Text.Render(m.remote))
		b.WriteString("\n\n")

		// Ahead/behind info
		if m.status.Info.Ahead > 0 {
			b.WriteString(m.theme.Success.Render(fmt.Sprintf("▲ %d commits to push", m.status.Info.Ahead)))
			b.WriteString("\n\n")

			// Show commits to be pushed
			b.WriteString(m.theme.Header.Render("Commits:"))
			b.WriteString("\n")
			for i, c := range m.commits {
				if i >= m.status.Info.Ahead {
					break
				}
				b.WriteString("  ")
				b.WriteString(m.theme.Muted.Render(c.Hash[:7]))
				b.WriteString(" ")
				b.WriteString(m.theme.Text.Render(c.Message))
				b.WriteString("\n")
			}
		} else if m.status.Info.Behind > 0 {
			b.WriteString(m.theme.Warning.Render(fmt.Sprintf("▼ %d commits behind remote", m.status.Info.Behind)))
			b.WriteString("\n")
			b.WriteString(m.theme.Muted.Render("Pull first to update your branch."))
		} else {
			b.WriteString(m.theme.Success.Render("Branch is up to date with remote"))
		}

		b.WriteString("\n\n")
		b.WriteString(m.theme.Help.Render("Enter Push   p Pull   r Refresh   b Back"))

	case pushConfirm:
		b.WriteString(m.theme.Header.Render("Confirm Push"))
		b.WriteString("\n\n")

		b.WriteString(m.theme.Text.Render(fmt.Sprintf("Push %d commits to %s/%s?",
			m.status.Info.Ahead, m.remote, m.status.Info.Branch)))
		b.WriteString("\n\n")

		// Options
		if m.forcePush {
			b.WriteString(m.theme.Selected.Render("[x]"))
		} else {
			b.WriteString(m.theme.Muted.Render("[ ]"))
		}
		b.WriteString(" Force push")
		b.WriteString("\n")

		if m.setUpstream {
			b.WriteString(m.theme.Selected.Render("[x]"))
		} else {
			b.WriteString(m.theme.Muted.Render("[ ]"))
		}
		b.WriteString(" Set upstream")
		b.WriteString("\n\n")

		b.WriteString(m.theme.Help.Render("Tab Toggle Force   u Upstream   Enter Confirm   Esc Cancel"))

	case pushing:
		b.WriteString(m.theme.Accent.Render("Pushing to remote..."))
		b.WriteString("\n\n")
		b.WriteString(m.theme.Muted.Render(fmt.Sprintf("%s/%s", m.remote, m.status.Info.Branch)))

	case pushDone:
		if m.err != nil {
			b.WriteString(m.theme.Error.Render("Push failed!"))
			b.WriteString("\n\n")
			b.WriteString(m.theme.Help.Render("r Retry   b Back"))
		} else {
			b.WriteString(m.theme.Success.Render("Push completed successfully!"))
			b.WriteString("\n\n")
			b.WriteString(m.theme.Help.Render("Enter Continue   b Back"))
		}
	}

	return lipgloss.NewStyle().Width(m.width).Height(m.height).Render(b.String())
}
