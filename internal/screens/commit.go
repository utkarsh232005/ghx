package screens

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/KDM-cli/ghx/internal/ai"
	"github.com/KDM-cli/ghx/internal/components"
	"github.com/KDM-cli/ghx/internal/db"
	"github.com/KDM-cli/ghx/internal/git"
	"github.com/KDM-cli/ghx/styles"
)

type commitState int

const (
	commitSelectFiles commitState = iota
	commitEnterMessage
	commitAISuggestions
	commitConfirming
	commitDone
)

type CommitModel struct {
	theme           *styles.Theme
	db              *db.DB
	aiManager       *ai.Manager
	state           commitState
	fileList        components.FileListModel
	message         components.TextInputModel
	suggestions     []string
	selectedSuggestion int
	loading         bool
	committing      bool
	width           int
	height          int
	err             error
	result          string
	spinner         spinner.Model
	generationStart time.Time
	elapsedTime     time.Duration
}

func NewCommitModel(theme *styles.Theme, database *db.DB, aiManager *ai.Manager) CommitModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = theme.Accent

	return CommitModel{
		theme:     theme,
		db:        database,
		aiManager: aiManager,
		state:     commitSelectFiles,
		message:   components.NewTextInputModel(theme, "Commit message..."),
		fileList:  components.NewFileListModel(theme, nil),
		spinner:   s,
	}
}

func (m CommitModel) Init() tea.Cmd {
	return tea.Batch(m.loadFiles, m.spinner.Tick)
}

func (m CommitModel) loadFiles() tea.Msg {
	status, err := git.GetStatus()
	if err != nil {
		return filesLoadedMsg{err: err}
	}

	var files []git.FileStatus
	files = append(files, status.Staged...)
	files = append(files, status.Modified...)
	files = append(files, status.Untracked...)

	return filesLoadedMsg{files: files}
}

type filesLoadedMsg struct {
	files []git.FileStatus
	err   error
}

type aiSuggestionsMsg struct {
	suggestions []string
	err         error
}

type commitResultMsg struct {
	success bool
	message string
	err     error
}

func (m CommitModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case spinner.TickMsg:
		var spinCmd tea.Cmd
		m.spinner, spinCmd = m.spinner.Update(msg)
		if m.loading {
			m.elapsedTime = time.Since(m.generationStart)
		}
		return m, spinCmd

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.message.SetWidth(m.width - 10)
		m.fileList.Height = m.height - 9
		m.fileList.Width = m.width

	case filesLoadedMsg:
		if msg.err != nil {
			m.err = msg.err
			return m, nil
		}
		m.fileList = m.fileList.SetFiles(msg.files)
		m.fileList.Height = m.height - 9
		m.fileList.Width = m.width
		m.state = commitSelectFiles

	case aiSuggestionsMsg:
		m.loading = false
		if msg.err != nil {
			m.err = msg.err
			m.state = commitEnterMessage
		} else {
			m.suggestions = msg.suggestions
			m.selectedSuggestion = 0
			m.state = commitAISuggestions
		}

	case commitResultMsg:
		m.committing = false
		if msg.success {
			m.state = commitDone
			m.result = msg.message
		} else {
			m.err = msg.err
			m.state = commitEnterMessage
		}

	case tea.KeyMsg:
		// Reset error on any key
		m.err = nil

		switch msg.String() {
		case "tab":
			switch m.state {
			case commitSelectFiles:
				if m.fileList.SelectedCount() > 0 {
					m.state = commitEnterMessage
				}
			case commitEnterMessage:
				m.state = commitSelectFiles
			case commitAISuggestions:
				m.state = commitEnterMessage
			}

		case "enter":
			switch m.state {
			case commitSelectFiles:
				if m.fileList.SelectedCount() > 0 {
					m.state = commitEnterMessage
				}
			case commitEnterMessage:
				if m.message.Value() != "" {
					m.committing = true
					return m, m.doCommit
				}
			case commitAISuggestions:
				if m.selectedSuggestion >= 0 && m.selectedSuggestion < len(m.suggestions) {
					m.message.SetValue(cleanSuggestion(m.suggestions[m.selectedSuggestion]))
					m.state = commitEnterMessage
				}
			case commitDone:
				// Reset for new commit
				m.state = commitSelectFiles
				m.message.SetValue("")
				m.suggestions = nil
				return m, m.loadFiles
			}

		case "g":
			if m.state == commitEnterMessage && m.fileList.SelectedCount() > 0 {
				m.loading = true
				m.generationStart = time.Now()
				m.elapsedTime = 0
				return m, m.generateAISuggestions
			}

		case "up", "k":
			if m.state == commitAISuggestions && m.selectedSuggestion > 0 {
				m.selectedSuggestion--
			}

		case "down", "j":
			if m.state == commitAISuggestions && m.selectedSuggestion < len(m.suggestions)-1 {
				m.selectedSuggestion++
			}

		case "1", "2", "3":
			if m.state == commitAISuggestions {
				idx := int(msg.String()[0] - '1')
				if idx >= 0 && idx < len(m.suggestions) {
					m.selectedSuggestion = idx
					m.message.SetValue(cleanSuggestion(m.suggestions[idx]))
					m.state = commitEnterMessage
				}
			}

		case "r":
			if m.state == commitAISuggestions || m.state == commitEnterMessage {
				m.loading = true
				m.generationStart = time.Now()
				m.elapsedTime = 0
				return m, m.generateAISuggestions
			}
		case "b":
			if m.state == commitSelectFiles || m.state == commitDone || m.state == commitAISuggestions {
				return m, func() tea.Msg {
					return Navigate(ScreenHome)
				}
			}
		}
	}

	// Update sub-components based on state
	switch m.state {
	case commitSelectFiles:
		m.fileList, cmd = m.fileList.Update(msg)
		cmds = append(cmds, cmd)
	case commitEnterMessage:
		m.message, cmd = m.message.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func cleanSuggestion(s string) string {
	// Remove leading numbers like "1. " or "1) "
	s = strings.TrimSpace(s)
	if len(s) > 3 && (s[1] == '.' || s[1] == ')' || s[1] == ' ') {
		s = strings.TrimSpace(s[2:])
	}
	return s
}

func (m CommitModel) generateAISuggestions() tea.Msg {
	paths := m.fileList.SelectedPaths()
	diff, err := git.GetDiff(paths)
	if err != nil {
		diff = ""
	}
	if diff == "" {
		// Try getting staged diff
		diff, _ = git.GetDiff(nil)
	}

	if diff == "" {
		return aiSuggestionsMsg{err: fmt.Errorf("no diff available")}
	}

	resp, err := m.aiManager.Chat(context.Background(), []ai.Message{
		{Role: "user", Content: ai.GenerateCommitMessagePrompt(diff)},
	})
	if err != nil {
		return aiSuggestionsMsg{err: err}
	}

	content := strings.TrimSpace(resp.Content)

	// Clean Markdown code block wrapper if present
	if strings.HasPrefix(content, "```json") {
		content = strings.TrimPrefix(content, "```json")
		content = strings.TrimSuffix(content, "```")
		content = strings.TrimSpace(content)
	} else if strings.HasPrefix(content, "```") {
		content = strings.TrimPrefix(content, "```")
		content = strings.TrimSuffix(content, "```")
		content = strings.TrimSpace(content)
	}

	// Try parsing as JSON array of strings
	var jsonSuggestions []string
	if err := json.Unmarshal([]byte(content), &jsonSuggestions); err == nil && len(jsonSuggestions) > 0 {
		var suggestions []string
		for i, s := range jsonSuggestions {
			suggestions = append(suggestions, fmt.Sprintf("%d. %s", i+1, strings.TrimSpace(s)))
		}
		return aiSuggestionsMsg{suggestions: suggestions}
	}

	// Fallback to parsing line-by-line
	lines := strings.Split(content, "\n")
	var suggestions []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Clean quote marks or brackets that some models wrap lines in
		line = strings.Trim(line, "`\"'{}[]")
		line = strings.TrimSpace(line)

		// Match lines starting with a number like "1. ", "2. ", "3. " or prefix numbers
		if len(line) > 2 && (line[0] >= '1' && line[0] <= '9') && (line[1] == '.' || line[1] == ')') {
			suggestions = append(suggestions, line)
		} else if strings.Contains(line, "feat(") || strings.Contains(line, "fix(") || strings.Contains(line, "refactor(") || strings.Contains(line, "style(") || strings.Contains(line, "docs(") || strings.Contains(line, "test(") || strings.Contains(line, "chore(") {
			suggestions = append(suggestions, fmt.Sprintf("%d. %s", len(suggestions)+1, line))
		}
	}

	// If no structured suggestions were extracted, fallback to raw lines
	if len(suggestions) == 0 {
		for i, line := range lines {
			if len(suggestions) >= 3 {
				break
			}
			line = strings.TrimSpace(line)
			if line != "" {
				suggestions = append(suggestions, fmt.Sprintf("%d. %s", i+1, line))
			}
		}
	}

	return aiSuggestionsMsg{suggestions: suggestions}
}

func (m CommitModel) doCommit() tea.Msg {
	paths := m.fileList.SelectedPaths()
	msg := m.message.Value()

	if err := git.CommitWithFiles(msg, paths); err != nil {
		return commitResultMsg{err: err}
	}

	return commitResultMsg{success: true, message: msg}
}

func (m CommitModel) View() string {
	var b strings.Builder

	if m.loading {
		b.WriteString(m.theme.Title.Render("Commit AI Suggestion"))
		b.WriteString("\n\n")

		loadingContent := fmt.Sprintf(
			"  %s  %s\n\n  %s\n\n  %s",
			m.spinner.View(),
			m.theme.Text.Bold(true).Render("Generating commit message suggestions..."),
			m.theme.Muted.Render("Analyzing selected changes and communicating with AI provider..."),
			m.theme.Accent.Render(fmt.Sprintf("Elapsed time: %.1fs", m.elapsedTime.Seconds())),
		)

		b.WriteString(m.theme.Box.Render(loadingContent))
		b.WriteString("\n\n")
		b.WriteString(m.theme.Help.Render("Please wait..."))

		return lipgloss.NewStyle().Width(m.width).Height(m.height).Render(b.String())
	}

	b.WriteString(m.theme.Title.Render("Commit"))
	b.WriteString("\n\n")

	if m.err != nil {
		b.WriteString(m.theme.Error.Render("Error: " + m.err.Error()))
		b.WriteString("\n\n")
	}

	switch m.state {
	case commitSelectFiles:
		b.WriteString(m.theme.Header.Render("Select files to commit"))
		b.WriteString("\n\n")

		if m.fileList.Count() == 0 {
			b.WriteString(m.theme.Muted.Render("No files to commit"))
			b.WriteString("\n")
		} else {
			b.WriteString(m.fileList.View())
		}
		b.WriteString("\n")
		b.WriteString(m.theme.Help.Render("Space Select   a All   n None   Tab/Enter Next"))

	case commitEnterMessage:
		b.WriteString(m.theme.Header.Render("Enter commit message"))
		b.WriteString("\n\n")

		count := m.fileList.SelectedCount()
		b.WriteString(m.theme.Text.Render(fmt.Sprintf("%d file", count)))
		if count != 1 {
			b.WriteString("s")
		}
		b.WriteString(m.theme.Muted.Render(" selected"))
		b.WriteString("\n\n")

		b.WriteString(m.message.View())
		b.WriteString("\n\n")

		if m.committing {
			b.WriteString(m.theme.Muted.Render("Committing..."))
		} else {
			b.WriteString(m.theme.Help.Render("g AI Generate   Tab Files   Enter Commit"))
		}

	case commitAISuggestions:
		b.WriteString(m.theme.Header.Render("AI Suggestions"))
		b.WriteString("\n\n")

		for i, s := range m.suggestions {
			prefix := "  "
			if i == m.selectedSuggestion {
				prefix = m.theme.Selected.Render("> ")
			}

			b.WriteString(prefix)
			b.WriteString(m.theme.Text.Render(fmt.Sprintf("%d. %s", i+1, cleanSuggestion(s))))
			b.WriteString("\n")
		}

		b.WriteString("\n")
		b.WriteString(m.theme.Help.Render("↑/↓ Navigate   1-3 Select   r Regenerate   Tab Edit   Enter Confirm"))

	case commitDone:
		b.WriteString(m.theme.Success.Render("Commit successful!"))
		b.WriteString("\n\n")
		b.WriteString(m.theme.Muted.Render(m.result))
		b.WriteString("\n\n")
		b.WriteString(m.theme.Help.Render("Enter New Commit   b Back"))
	}

	return lipgloss.NewStyle().Width(m.width).Height(m.height).Render(b.String())
}
