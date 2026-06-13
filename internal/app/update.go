package app

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/utkarshpatrikar/ghx/internal/ai"
	"github.com/utkarshpatrikar/ghx/internal/git"
)

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.state.Width = msg.Width
		m.state.Height = msg.Height
		return m, nil
	case configLoadedMsg:
		if msg.err != nil {
			m.state.Message = fmt.Sprintf("Using default config: %v", msg.err)
		} else {
			m.state.Config = msg.config
		}
		return m, nil
	case statusLoadedMsg:
		m.state.Loading = false
		m.state.Err = msg.err
		m.state.StatusFiles = msg.files
		m.state.RepoInfo = msg.info
		m.ensureSelection()
		if m.state.FileCursor >= len(m.state.StatusFiles) {
			m.state.FileCursor = max(0, len(m.state.StatusFiles)-1)
		}
		if msg.err != nil {
			m.state.Message = msg.err.Error()
		} else if len(msg.files) == 0 {
			m.state.Message = "Working tree clean"
		} else {
			m.state.Message = fmt.Sprintf("%d changed file(s)", len(msg.files))
		}
		if m.state.Screen == ScreenDiff {
			m.state.Loading = true
			return m, loadDiff(m.currentPathSlice())
		}
		return m, nil
	case diffLoadedMsg:
		m.state.Loading = false
		m.state.Err = msg.err
		if msg.err != nil {
			m.state.Message = msg.err.Error()
			m.state.DiffText = ""
		} else if strings.TrimSpace(msg.text) == "" {
			m.state.Message = "No diff for selected file"
			m.state.DiffText = "No diff for selected file."
		} else {
			m.state.Message = "Diff loaded"
			m.state.DiffText = msg.text
		}
		return m, nil
	case commandFinishedMsg:
		m.state.Loading = false
		m.state.PushConfirm = false
		m.state.Output = strings.TrimSpace(msg.output)
		if msg.err != nil {
			m.state.Err = msg.err
			m.state.Message = msg.err.Error()
			return m, nil
		}
		m.state.Err = nil
		m.state.Message = "Command completed"
		m.state.BranchInputActive = false
		return m, loadStatus
	case issuesLoadedMsg:
		m.state.Loading = false
		m.state.Err = msg.err
		if msg.err != nil {
			m.state.Message = msg.err.Error()
			m.state.Issues = nil
		} else {
			m.state.Message = fmt.Sprintf("%d issue(s) loaded", len(msg.issues))
			m.state.Issues = msg.issues
		}
		return m, nil
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "q":
			if m.textInputActive() {
				break
			}
			if m.state.Screen == ScreenHome {
				return m, tea.Quit
			}
			m.state.Screen = ScreenHome
			m.state.Err = nil
			m.state.Message = ""
			return m, nil
		case "esc":
			if m.state.Screen == ScreenPR && m.state.BranchInputActive {
				m.state.BranchInputActive = false
				m.state.NewBranchName = ""
				m.state.Err = nil
				m.state.Message = ""
				return m, nil
			}
			if m.state.Screen == ScreenHome {
				return m, tea.Quit
			}
			m.state.Screen = ScreenHome
			m.state.Err = nil
			m.state.Message = ""
			return m, nil
		}

		if m.state.Screen == ScreenHome {
			return updateHome(m, msg)
		}
		return updateScreen(m, msg)
	}

	return m, nil
}

func updateHome(m Model, msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		m.state.Menu.Prev()
	case "down", "j":
		m.state.Menu.Next()
	case "enter":
		item := m.state.Menu.Selected()
		switch item.Target {
		case "status":
			m.state.Screen = ScreenStatus
			m.state.Loading = true
			m.state.Message = "Loading git status..."
			return m, loadStatus
		case "commit":
			return openWorkflow(m, ScreenCommit, "Loading commit workflow...")
		case "diff":
			return openWorkflow(m, ScreenDiff, "Loading diff workflow...")
		case "push":
			return openWorkflow(m, ScreenPush, "Loading push workflow...")
		case "pr":
			return openWorkflow(m, ScreenPR, "Loading PR workflow...")
		case "issues":
			m.state.Screen = ScreenIssues
			m.state.Loading = true
			m.state.Message = "Loading issues..."
			return m, loadIssues
		case "repos":
			return openWorkflow(m, ScreenRepos, "Loading repo details...")
		case "ai":
			return openWorkflow(m, ScreenAI, "Loading AI assistant...")
		case "history":
			return openWorkflow(m, ScreenHistory, "Loading history...")
		case "settings":
			m.state.Screen = ScreenSettings
			m.state.Message = "Settings loaded"
		default:
			m.state.Message = fmt.Sprintf("%s is not wired yet", item.Title)
		}
	}

	return m, nil
}

func openWorkflow(m Model, screen Screen, message string) (tea.Model, tea.Cmd) {
	m.state.Screen = screen
	m.state.Loading = true
	m.state.Message = message
	return m, loadStatus
}

func updateScreen(m Model, msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.state.Screen {
	case ScreenCommit:
		return updateCommit(m, msg)
	case ScreenDiff:
		return updateDiff(m, msg)
	case ScreenPush:
		return updatePush(m, msg)
	case ScreenPR:
		return updatePR(m, msg)
	case ScreenIssues:
		return updateIssues(m, msg)
	case ScreenAI:
		return updateAI(m, msg)
	case ScreenSettings:
		return updateSettings(m, msg)
	}

	switch msg.String() {
	case "r":
		if m.state.Screen == ScreenStatus || m.state.Screen == ScreenRepos {
			m.state.Loading = true
			m.state.Message = "Refreshing..."
			return m, loadStatus
		}
	case "b":
		m.state.Screen = ScreenHome
		m.state.Message = ""
	}
	return m, nil
}

func updateCommit(m Model, msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if m.state.CommitFocus == 0 {
			m.moveFileCursor(-1)
		}
	case "down", "j":
		if m.state.CommitFocus == 0 {
			m.moveFileCursor(1)
		}
	case "tab":
		m.state.CommitFocus = (m.state.CommitFocus + 1) % 2
	case "shift+tab":
		m.state.CommitFocus = (m.state.CommitFocus + 1) % 2
	case " ":
		if m.state.CommitFocus == 0 {
			m.toggleCurrentFile()
		} else {
			m.state.CommitMessage += " "
		}
	case "ctrl+u":
		m.state.CommitMessage = ""
	case "backspace":
		if m.state.CommitFocus == 1 && len(m.state.CommitMessage) > 0 {
			m.state.CommitMessage = m.state.CommitMessage[:len(m.state.CommitMessage)-1]
		}
	case "g":
		paths := m.selectedPaths()
		m.state.CommitMessage = git.SuggestedCommitMessage(m.state.StatusFiles, paths)
		m.state.Message = "Generated commit message"
	case "enter":
		if m.state.CommitFocus == 0 {
			m.state.CommitFocus = 1
			return m, nil
		}
		paths := m.selectedPaths()
		m.state.Loading = true
		m.state.Message = "Committing selected files..."
		return m, runCommit(paths, m.state.CommitMessage)
	case "r":
		m.state.Loading = true
		m.state.Message = "Refreshing commit workflow..."
		return m, loadStatus
	case "b":
		m.state.Screen = ScreenHome
		m.state.Message = ""
	default:
		if m.state.CommitFocus == 1 && len(msg.Runes) > 0 {
			m.state.CommitMessage += string(msg.Runes)
		}
	}
	return m, nil
}

func updateDiff(m Model, msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		m.moveFileCursor(-1)
		m.state.DiffScroll = 0
		return m, loadDiff(m.currentPathSlice())
	case "down", "j":
		m.moveFileCursor(1)
		m.state.DiffScroll = 0
		return m, loadDiff(m.currentPathSlice())
	case "pgup":
		m.state.DiffScroll = max(0, m.state.DiffScroll-10)
	case "pgdown":
		m.state.DiffScroll += 10
	case "r", "enter":
		m.state.Loading = true
		m.state.Message = "Loading diff..."
		return m, loadDiff(m.currentPathSlice())
	case "b":
		m.state.Screen = ScreenHome
		m.state.Message = ""
	}
	return m, nil
}

func updatePush(m Model, msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "p", "enter":
		if !m.state.PushConfirm {
			m.state.PushConfirm = true
			m.state.Message = "Press p again to push"
			return m, nil
		}
		m.state.Loading = true
		m.state.Message = "Pushing..."
		return m, runPush
	case "r":
		m.state.Loading = true
		m.state.Message = "Refreshing push workflow..."
		return m, loadStatus
	case "b":
		m.state.Screen = ScreenHome
		m.state.Message = ""
	}
	return m, nil
}

func updatePR(m Model, msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "tab":
		if !m.state.BranchInputActive {
			m.state.PRFocus = (m.state.PRFocus + 1) % 2
		}
	case "shift+tab":
		if !m.state.BranchInputActive {
			m.state.PRFocus = (m.state.PRFocus + 1) % 2
		}
	case "ctrl+u":
		if m.state.BranchInputActive {
			m.state.NewBranchName = ""
		} else if m.state.PRFocus == 0 {
			m.state.PRTitle = ""
		} else {
			m.state.PRBody = ""
		}
	case "backspace":
		if m.state.BranchInputActive {
			if len(m.state.NewBranchName) > 0 {
				m.state.NewBranchName = m.state.NewBranchName[:len(m.state.NewBranchName)-1]
			}
		} else {
			if m.state.PRFocus == 0 && len(m.state.PRTitle) > 0 {
				m.state.PRTitle = m.state.PRTitle[:len(m.state.PRTitle)-1]
			}
			if m.state.PRFocus == 1 && len(m.state.PRBody) > 0 {
				m.state.PRBody = m.state.PRBody[:len(m.state.PRBody)-1]
			}
		}
	case "g":
		if !m.state.BranchInputActive {
			m.state.PRTitle = git.SuggestedCommitMessage(m.state.StatusFiles, m.allPaths())
			m.state.PRBody = m.generatedPRBody()
			m.state.Message = "Generated PR draft"
		} else {
			if len(msg.Runes) > 0 {
				m.state.NewBranchName += string(msg.Runes)
			}
		}
	case "c":
		if !m.state.BranchInputActive {
			m.state.BranchInputActive = true
			m.state.NewBranchName = ""
		} else {
			if len(msg.Runes) > 0 {
				m.state.NewBranchName += string(msg.Runes)
			}
		}
	case "enter":
		if m.state.BranchInputActive {
			name := strings.TrimSpace(m.state.NewBranchName)
			if name == "" {
				m.state.Message = "Branch name cannot be empty"
				return m, nil
			}
			m.state.Loading = true
			m.state.Message = fmt.Sprintf("Checking out branch %s...", name)
			return m, runCheckoutBranch(name)
		}
		m.state.Loading = true
		m.state.Message = "Creating pull request..."
		return m, runCreatePR(m.state.PRTitle, m.state.PRBody)
	case "r":
		if !m.state.BranchInputActive {
			m.state.Loading = true
			m.state.Message = "Refreshing PR workflow..."
			return m, loadStatus
		} else {
			if len(msg.Runes) > 0 {
				m.state.NewBranchName += string(msg.Runes)
			}
		}
	case "b":
		if !m.state.BranchInputActive {
			m.state.Screen = ScreenHome
			m.state.Message = ""
		} else {
			if len(msg.Runes) > 0 {
				m.state.NewBranchName += string(msg.Runes)
			}
		}
	default:
		if len(msg.Runes) > 0 {
			if m.state.BranchInputActive {
				m.state.NewBranchName += string(msg.Runes)
			} else if m.state.PRFocus == 0 {
				m.state.PRTitle += string(msg.Runes)
			} else {
				m.state.PRBody += string(msg.Runes)
			}
		}
	}
	return m, nil
}

func updateIssues(m Model, msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "r", "enter":
		m.state.Loading = true
		m.state.Message = "Loading issues..."
		return m, loadIssues
	case "b":
		m.state.Screen = ScreenHome
		m.state.Message = ""
	}
	return m, nil
}

func updateAI(m Model, msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		if strings.TrimSpace(m.state.AIInput) == "" {
			m.state.Message = "Ask something first"
			return m, nil
		}
		m.state.AIResponse = m.localAIResponse()
		m.state.Message = "Answered locally"
	case "ctrl+u":
		m.state.AIInput = ""
	case "backspace":
		if len(m.state.AIInput) > 0 {
			m.state.AIInput = m.state.AIInput[:len(m.state.AIInput)-1]
		}
	case "b":
		m.state.Screen = ScreenHome
		m.state.Message = ""
	default:
		if len(msg.Runes) > 0 {
			m.state.AIInput += string(msg.Runes)
		}
	}
	return m, nil
}

func updateSettings(m Model, msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "n":
		m.state.Config.AI.ActiveProvider = nextProvider(m.state.Config.AI.ActiveProvider)
		m.state.Message = "Provider changed"
	case "s":
		if err := ai.SaveConfig(m.state.Config); err != nil {
			m.state.Err = err
			m.state.Message = err.Error()
		} else {
			m.state.Err = nil
			m.state.Message = "Saved .ghx/config.json"
		}
	case "b":
		m.state.Screen = ScreenHome
		m.state.Message = ""
	}
	return m, nil
}

func (m *Model) ensureSelection() {
	if m.state.SelectedFiles == nil {
		m.state.SelectedFiles = map[string]bool{}
	}
	for _, file := range m.state.StatusFiles {
		if _, ok := m.state.SelectedFiles[file.Path]; !ok {
			m.state.SelectedFiles[file.Path] = true
		}
	}
}

func (m *Model) moveFileCursor(delta int) {
	if len(m.state.StatusFiles) == 0 {
		m.state.FileCursor = 0
		return
	}
	m.state.FileCursor += delta
	if m.state.FileCursor < 0 {
		m.state.FileCursor = len(m.state.StatusFiles) - 1
	}
	if m.state.FileCursor >= len(m.state.StatusFiles) {
		m.state.FileCursor = 0
	}
}

func (m *Model) toggleCurrentFile() {
	if len(m.state.StatusFiles) == 0 {
		return
	}
	m.ensureSelection()
	path := m.state.StatusFiles[m.state.FileCursor].Path
	m.state.SelectedFiles[path] = !m.state.SelectedFiles[path]
}

func (m Model) selectedPaths() []string {
	paths := []string{}
	for _, file := range m.state.StatusFiles {
		if m.state.SelectedFiles[file.Path] {
			paths = append(paths, file.Path)
		}
	}
	return paths
}

func (m Model) allPaths() []string {
	paths := make([]string, 0, len(m.state.StatusFiles))
	for _, file := range m.state.StatusFiles {
		paths = append(paths, file.Path)
	}
	return paths
}

func (m Model) currentPathSlice() []string {
	if len(m.state.StatusFiles) == 0 {
		return nil
	}
	return []string{m.state.StatusFiles[m.state.FileCursor].Path}
}

func (m Model) generatedPRBody() string {
	lines := []string{"## Summary", "", "- Update repository files.", "", "## Changes"}
	for _, file := range m.state.StatusFiles {
		lines = append(lines, "- "+file.ShortStatus()+" "+file.Path)
	}
	lines = append(lines, "", "## Testing", "", "- Not run yet.")
	return strings.Join(lines, "\n")
}

func (m Model) localAIResponse() string {
	question := strings.ToLower(m.state.AIInput)
	if strings.Contains(question, "commit") {
		return "Suggested commit message: " + git.SuggestedCommitMessage(m.state.StatusFiles, m.selectedPaths())
	}
	if strings.Contains(question, "status") || strings.Contains(question, "change") {
		return m.statusSummaryText()
	}
	if strings.Contains(question, "branch") {
		return "Current branch: " + m.state.RepoInfo.Branch
	}
	return "I can answer from local repo context right now. Try asking about status, branch, or commit message suggestions."
}

func (m Model) statusSummaryText() string {
	if len(m.state.StatusFiles) == 0 {
		return "Working tree clean."
	}
	lines := []string{fmt.Sprintf("%d changed file(s):", len(m.state.StatusFiles))}
	for _, file := range m.state.StatusFiles {
		lines = append(lines, file.ShortStatus()+" "+file.Path)
	}
	return strings.Join(lines, "\n")
}

func nextProvider(current string) string {
	order := []string{"ollama", "openai", "claude", "lmstudio", "mlx"}
	for i, provider := range order {
		if provider == current {
			return order[(i+1)%len(order)]
		}
	}
	return order[0]
}

func max(a int, b int) int {
	if a > b {
		return a
	}
	return b
}

func (m Model) textInputActive() bool {
	return (m.state.Screen == ScreenCommit && m.state.CommitFocus == 1) ||
		(m.state.Screen == ScreenPR) ||
		(m.state.Screen == ScreenAI)
}
