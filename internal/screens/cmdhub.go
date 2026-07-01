package screens

import (
	"context"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/KDM-cli/ghx/internal/ai"
	"github.com/KDM-cli/ghx/internal/commands"
	"github.com/KDM-cli/ghx/internal/db"
	"github.com/KDM-cli/ghx/styles"
)

type cmdHubState int

const (
	cmdHubSelectGroup cmdHubState = iota
	cmdHubSelectCmd
	cmdHubConfigure
	cmdHubExecuting
	cmdHubExecFinished
	cmdHubConfirmDestructive
	cmdHubExplainAI
)

type CmdHubModel struct {
	theme            *styles.Theme
	db               *db.DB
	aiManager        *ai.Manager
	state            cmdHubState
	catalog          []commands.CommandGroup
	
	// Selection state
	selectedGroupIdx int
	selectedCmdIdx   int
	searchField      textinput.Model
	searchFocused    bool
	filteredCmds     []filteredCmdItem

	// Configuration state
	selectedCmd      commands.CommandDef
	paramValues      map[string]string
	focusedParamIdx  int // 0 to len(params), where len(params) is the "Run" button
	paramInput       textinput.Model

	// Execution state
	spinner          spinner.Model
	loading          bool
	runResult        commands.RunResult
	scrollOffset     int
	aiExplainText    string
	aiLoading        bool
	previousState    cmdHubState

	// Window dimensions
	width            int
	height           int
}

type filteredCmdItem struct {
	groupIndex int
	cmdIndex   int
	cmdDef     commands.CommandDef
}

type runFinishedMsg struct {
	result commands.RunResult
}

type aiExplainFinishedMsg struct {
	explanation string
	err         error
}

func NewCmdHubModel(theme *styles.Theme, database *db.DB, aiManager *ai.Manager) CmdHubModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = theme.Accent

	sf := textinput.New()
	sf.Placeholder = "Search commands... (Press '/' to search)"
	sf.Prompt = " / "
	sf.PromptStyle = theme.Accent

	pi := textinput.New()
	pi.Prompt = " Value: "
	pi.PromptStyle = theme.Accent

	return CmdHubModel{
		theme:         theme,
		db:            database,
		aiManager:     aiManager,
		state:         cmdHubSelectGroup,
		catalog:       commands.GetCatalog(),
		searchField:   sf,
		paramInput:    pi,
		spinner:       s,
		paramValues:   make(map[string]string),
	}
}

func (m CmdHubModel) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
	)
}

func (m CmdHubModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.searchField.Width = m.width - 10
		m.paramInput.Width = m.width - 25

	case spinner.TickMsg:
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case runFinishedMsg:
		m.loading = false
		m.runResult = msg.result
		m.scrollOffset = 0
		m.state = cmdHubExecFinished
		return m, nil

	case aiExplainFinishedMsg:
		m.aiLoading = false
		if msg.err != nil {
			m.aiExplainText = "Error communicating with AI: " + msg.err.Error()
		} else {
			m.aiExplainText = msg.explanation
		}
		m.scrollOffset = 0
		return m, nil

	case tea.KeyMsg:
		// Reset/handle search focus
		if m.searchFocused {
			switch msg.String() {
			case "enter", "esc":
				m.searchFocused = false
				m.searchField.Blur()
				m.updateFiltering()
				return m, nil
			default:
				m.searchField, cmd = m.searchField.Update(msg)
				m.updateFiltering()
				return m, cmd
			}
		}

		// Handle keys based on TUI sub-state
		switch m.state {
		case cmdHubSelectGroup:
			switch msg.String() {
			case "/":
				m.searchFocused = true
				m.searchField.Focus()
				m.searchField.SetValue("")
				m.updateFiltering()
				return m, nil
			case "up", "k":
				if m.selectedGroupIdx > 0 {
					m.selectedGroupIdx--
				}
			case "down", "j":
				if m.selectedGroupIdx < len(m.catalog)-1 {
					m.selectedGroupIdx++
				}
			case "enter", "right", "l":
				if len(m.catalog) > 0 && len(m.catalog[m.selectedGroupIdx].Commands) > 0 {
					m.state = cmdHubSelectCmd
					m.selectedCmdIdx = 0
				}
			case "b", "esc":
				return m, func() tea.Msg {
					return Navigate(ScreenHome)
				}
			}

		case cmdHubSelectCmd:
			switch msg.String() {
			case "/":
				m.searchFocused = true
				m.searchField.Focus()
				m.searchField.SetValue("")
				m.updateFiltering()
				return m, nil
			case "up", "k":
				if m.isFiltering() {
					if m.selectedCmdIdx > 0 {
						m.selectedCmdIdx--
					}
				} else {
					if m.selectedCmdIdx > 0 {
						m.selectedCmdIdx--
					}
				}
			case "down", "j":
				if m.isFiltering() {
					if m.selectedCmdIdx < len(m.filteredCmds)-1 {
						m.selectedCmdIdx++
					}
				} else {
					if m.selectedCmdIdx < len(m.catalog[m.selectedGroupIdx].Commands)-1 {
						m.selectedCmdIdx++
					}
				}
			case "esc", "left", "h":
				if m.isFiltering() {
					m.searchField.SetValue("")
					m.updateFiltering()
				}
				m.state = cmdHubSelectGroup
			case "enter":
				var targetCmd commands.CommandDef
				if m.isFiltering() {
					if len(m.filteredCmds) > 0 && m.selectedCmdIdx < len(m.filteredCmds) {
						targetCmd = m.filteredCmds[m.selectedCmdIdx].cmdDef
					} else {
						return m, nil
					}
				} else {
					group := m.catalog[m.selectedGroupIdx]
					if len(group.Commands) > 0 && m.selectedCmdIdx < len(group.Commands) {
						targetCmd = group.Commands[m.selectedCmdIdx]
					} else {
						return m, nil
					}
				}

				m.selectedCmd = targetCmd
				m.paramValues = make(map[string]string)
				for _, p := range m.selectedCmd.Parameters {
					m.paramValues[p.Name] = p.DefaultValue
				}
				m.focusedParamIdx = 0
				m.state = cmdHubConfigure
				m.syncParamInput()
			}

		case cmdHubConfigure:
			// If we are currently editing a ParamString parameter, direct keys to text input
			isEditingString := false
			var activeParam commands.Parameter
			if m.focusedParamIdx < len(m.selectedCmd.Parameters) {
				activeParam = m.selectedCmd.Parameters[m.focusedParamIdx]
				if activeParam.Type == commands.ParamString {
					isEditingString = true
				}
			}

			if isEditingString && msg.String() != "enter" && msg.String() != "esc" && msg.String() != "up" && msg.String() != "down" && msg.String() != "tab" {
				m.paramInput, cmd = m.paramInput.Update(msg)
				m.paramValues[activeParam.Name] = m.paramInput.Value()
				return m, cmd
			}

			switch msg.String() {
			case "up", "k":
				if m.focusedParamIdx > 0 {
					m.focusedParamIdx--
					m.syncParamInput()
				}
			case "down", "j", "tab":
				if m.focusedParamIdx < len(m.selectedCmd.Parameters) {
					m.focusedParamIdx++
					m.syncParamInput()
				}
			case "space":
				if m.focusedParamIdx < len(m.selectedCmd.Parameters) {
					param := m.selectedCmd.Parameters[m.focusedParamIdx]
					if param.Type == commands.ParamBool {
						if m.paramValues[param.Name] == "true" {
							m.paramValues[param.Name] = "false"
						} else {
							m.paramValues[param.Name] = "true"
						}
					} else if param.Type == commands.ParamChoice {
						// Cycle choices forward
						idx := -1
						currVal := m.paramValues[param.Name]
						for i, choice := range param.Choices {
							if choice == currVal {
								idx = i
								break
							}
						}
						nextIdx := (idx + 1) % len(param.Choices)
						m.paramValues[param.Name] = param.Choices[nextIdx]
					}
				}
			case "left", "h":
				if m.focusedParamIdx < len(m.selectedCmd.Parameters) {
					param := m.selectedCmd.Parameters[m.focusedParamIdx]
					if param.Type == commands.ParamChoice {
						// Cycle choices backward
						idx := -1
						currVal := m.paramValues[param.Name]
						for i, choice := range param.Choices {
							if choice == currVal {
								idx = i
								break
							}
						}
						prevIdx := idx - 1
						if prevIdx < 0 {
							prevIdx = len(param.Choices) - 1
						}
						m.paramValues[param.Name] = param.Choices[prevIdx]
					}
				}
			case "enter":
				// Run the command
				if m.selectedCmd.Destructive {
					m.state = cmdHubConfirmDestructive
				} else {
					return m, m.triggerExecution()
				}
			case "ctrl+g":
				// Ask AI to explain command
				m.state = cmdHubExplainAI
				m.previousState = cmdHubConfigure
				m.aiLoading = true
				m.aiExplainText = ""
				return m, m.generateAIExplanation(false)
			case "esc", "b":
				m.state = cmdHubSelectCmd
			}

		case cmdHubConfirmDestructive:
			switch msg.String() {
			case "y", "Y":
				return m, m.triggerExecution()
			case "n", "N", "esc":
				m.state = cmdHubConfigure
			}

		case cmdHubExecFinished:
			switch msg.String() {
			case "up", "k":
				if m.scrollOffset > 0 {
					m.scrollOffset--
				}
			case "down", "j":
				lines := strings.Split(m.runResult.Output, "\n")
				maxScroll := len(lines) - (m.height - 10)
				if maxScroll < 0 {
					maxScroll = 0
				}
				if m.scrollOffset < maxScroll {
					m.scrollOffset++
				}
			case "r":
				// Rerun command
				return m, m.triggerExecution()
			case "ctrl+g":
				// Ask AI to explain error/output
				m.state = cmdHubExplainAI
				m.previousState = cmdHubExecFinished
				m.aiLoading = true
				m.aiExplainText = ""
				isError := !m.runResult.Success
				return m, m.generateAIExplanation(isError)
			case "esc", "b":
				m.state = cmdHubConfigure
			}

		case cmdHubExplainAI:
			switch msg.String() {
			case "up", "k":
				if m.scrollOffset > 0 {
					m.scrollOffset--
				}
			case "down", "j":
				lines := strings.Split(m.aiExplainText, "\n")
				maxScroll := len(lines) - (m.height - 10)
				if maxScroll < 0 {
					maxScroll = 0
				}
				if m.scrollOffset < maxScroll {
					m.scrollOffset++
				}
			case "esc", "b":
				m.state = m.previousState
			}
		}
	}

	return m, tea.Batch(cmds...)
}

func (m *CmdHubModel) syncParamInput() {
	if m.focusedParamIdx < len(m.selectedCmd.Parameters) {
		param := m.selectedCmd.Parameters[m.focusedParamIdx]
		if param.Type == commands.ParamString {
			m.paramInput.Focus()
			m.paramInput.SetValue(m.paramValues[param.Name])
		} else {
			m.paramInput.Blur()
		}
	} else {
		m.paramInput.Blur()
	}
}

func (m CmdHubModel) isFiltering() bool {
	return m.searchField.Value() != ""
}

func (m *CmdHubModel) updateFiltering() {
	query := strings.ToLower(m.searchField.Value())
	m.filteredCmds = nil
	if query == "" {
		return
	}

	for gIdx, group := range m.catalog {
		for cIdx, cmd := range group.Commands {
			if strings.Contains(strings.ToLower(cmd.Name), query) || strings.Contains(strings.ToLower(cmd.Description), query) {
				m.filteredCmds = append(m.filteredCmds, filteredCmdItem{
					groupIndex: gIdx,
					cmdIndex:   cIdx,
					cmdDef:     cmd,
				})
			}
		}
	}
	m.state = cmdHubSelectCmd
	m.selectedCmdIdx = 0
}

func (m CmdHubModel) triggerExecution() tea.Cmd {
	if m.selectedCmd.RequiresSuspend {
		return m.runInteractive()
	}

	m.loading = true
	m.state = cmdHubExecuting
	return func() tea.Msg {
		res := commands.RunCommand(context.Background(), m.db, m.selectedCmd, m.paramValues)
		return runFinishedMsg{result: res}
	}
}

func (m CmdHubModel) runInteractive() tea.Cmd {
	return commands.SuspendAndRun(m.db, m.selectedCmd, m.paramValues, func(err error) tea.Msg {
		var res commands.RunResult
		res.FullCmd = m.selectedCmd.CommandBase + " " + strings.Join(commands.BuildArgs(m.selectedCmd, m.paramValues), " ")
		if err != nil {
			res.Error = err
			res.Output = fmt.Sprintf("Interactive session finished with error: %v", err)
			res.Success = false
		} else {
			res.Output = "Interactive session completed successfully."
			res.Success = true
		}
		return runFinishedMsg{result: res}
	})
}

func (m CmdHubModel) generateAIExplanation(isError bool) tea.Cmd {
	return func() tea.Msg {
		var prompt string
		cmdStr := m.selectedCmd.CommandBase + " " + strings.Join(commands.BuildArgs(m.selectedCmd, m.paramValues), " ")
		if isError {
			prompt = ai.GenerateCommandErrorPrompt(cmdStr, m.runResult.Output)
		} else {
			prompt = ai.GenerateCommandExplanationPrompt(cmdStr)
		}

		resp, err := m.aiManager.Chat(context.Background(), []ai.Message{
			{Role: "user", Content: prompt},
		})
		if err != nil {
			return aiExplainFinishedMsg{err: err}
		}
		return aiExplainFinishedMsg{explanation: resp.Content}
	}
}

func (m CmdHubModel) View() string {
	var b strings.Builder

	// Top Title banner
	b.WriteString(m.theme.Title.Render("Git & GitHub CLI Hub"))
	b.WriteString("\n\n")

	switch m.state {
	case cmdHubSelectGroup, cmdHubSelectCmd:
		// Search Field
		if m.searchFocused {
			b.WriteString(m.theme.Selected.Render("Searching: " + m.searchField.View()))
		} else {
			b.WriteString(m.theme.Muted.Render(m.searchField.View()))
		}
		b.WriteString("\n\n")

		if m.isFiltering() {
			b.WriteString(m.theme.Header.Render(fmt.Sprintf("Search Results (%d match(es))", len(m.filteredCmds))))
			b.WriteString("\n\n")
			if len(m.filteredCmds) == 0 {
				b.WriteString(m.theme.Muted.Render("  No commands found matching query."))
			} else {
				visibleCount := m.height - 10
				if visibleCount < 3 {
					visibleCount = 3
				}
				start, end := getViewportRange(m.selectedCmdIdx, len(m.filteredCmds), visibleCount)
				for i := start; i < end; i++ {
					item := m.filteredCmds[i]
					prefix := "  "
					if i == m.selectedCmdIdx {
						prefix = m.theme.Selected.Render("> ")
					}
					b.WriteString(prefix)
					b.WriteString(m.theme.Text.Render(item.cmdDef.Name))
					b.WriteString("  " + m.theme.Muted.Render(item.cmdDef.Description))
					b.WriteString("\n")
				}
			}
			b.WriteString("\n")
			b.WriteString(m.theme.Help.Render("/ Search   ↑/↓ Select   Enter Configure   esc Clear Search   b Back"))
		} else {
			// Two-column layout
			leftWidth := 25
			rightWidth := m.width - leftWidth - 5
			if rightWidth < 30 {
				rightWidth = 30
			}

			// Render Left side (Groups)
			var leftSide strings.Builder
			leftSide.WriteString(m.theme.Header.Render("Categories"))
			leftSide.WriteString("\n\n")
			for i, group := range m.catalog {
				if i == m.selectedGroupIdx && m.state == cmdHubSelectGroup {
					leftSide.WriteString(m.theme.Selected.Render("> " + group.Name) + "\n")
				} else if i == m.selectedGroupIdx {
					leftSide.WriteString(m.theme.Text.Bold(true).Render("  " + group.Name) + "\n")
				} else {
					leftSide.WriteString(m.theme.Muted.Render("  " + group.Name) + "\n")
				}
			}

			// Render Right side (Commands under selected group)
			var rightSide strings.Builder
			group := m.catalog[m.selectedGroupIdx]
			rightSide.WriteString(m.theme.Header.Render(group.Name))
			rightSide.WriteString("\n\n")
			for i, cmd := range group.Commands {
				prefix := "  "
				if i == m.selectedCmdIdx && m.state == cmdHubSelectCmd {
					prefix = m.theme.Selected.Render("> ")
				}
				rightSide.WriteString(prefix)
				rightSide.WriteString(m.theme.Text.Render(cmd.Name))
				rightSide.WriteString("\n    " + m.theme.Muted.Render(cmd.Description))
				rightSide.WriteString("\n\n")
			}

			leftBox := lipgloss.NewStyle().Width(leftWidth).Render(leftSide.String())
			rightBox := lipgloss.NewStyle().Width(rightWidth).Render(rightSide.String())

			b.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, leftBox, "   ", rightBox))
			b.WriteString("\n")

			if m.state == cmdHubSelectGroup {
				b.WriteString(m.theme.Help.Render("/ Search   ↑/↓ Select   Enter Choose Group   b Back"))
			} else {
				b.WriteString(m.theme.Help.Render("/ Search   ↑/↓ Select   Enter Configure   esc Back to Groups"))
			}
		}

	case cmdHubConfigure:
		cmdDef := m.selectedCmd
		b.WriteString(m.theme.Header.Render("Configure parameters: " + cmdDef.Name))
		b.WriteString("\n")
		b.WriteString(m.theme.Muted.Render(cmdDef.Description))
		b.WriteString("\n\n")

		// List of parameters
		for i, param := range cmdDef.Parameters {
			prefix := "  "
			if i == m.focusedParamIdx {
				prefix = m.theme.Selected.Render("> ")
			}

			b.WriteString(prefix)
			b.WriteString(m.theme.Bold.Render(param.Name))
			b.WriteString(": ")

			val := m.paramValues[param.Name]
			if param.Type == commands.ParamBool {
				// Style boolean output nicely
				checked := val == "true"
				b.WriteString(m.theme.Checkbox(checked))
				b.WriteString("  " + m.theme.Muted.Render("(Press Space to toggle)"))
			} else if param.Type == commands.ParamChoice {
				b.WriteString(m.theme.Accent.Render("[" + val + "]"))
				b.WriteString("  " + m.theme.Muted.Render("(Press Space or Left/Right to cycle choices)"))
			} else {
				// String textinput parameter
				if i == m.focusedParamIdx {
					b.WriteString(m.paramInput.View())
				} else {
					if val == "" {
						b.WriteString(m.theme.Muted.Render("[empty]"))
					} else {
						b.WriteString(m.theme.Text.Render(val))
					}
				}
			}
			b.WriteString("\n")
			b.WriteString(m.theme.Help.Render("    " + param.Description))
			b.WriteString("\n\n")
		}

		// Run Button
		runPrefix := "  "
		if m.focusedParamIdx == len(cmdDef.Parameters) {
			runPrefix = m.theme.Selected.Render("> ")
		}
		b.WriteString(runPrefix)
		if cmdDef.Destructive {
			b.WriteString(m.theme.Error.Render("[ Run Command (Destructive) ]"))
		} else {
			b.WriteString(m.theme.Success.Render("[ Run Command ]"))
		}
		b.WriteString("\n\n")

		// Render Live Preview
		args := commands.BuildArgs(cmdDef, m.paramValues)
		fullCmdStr := cmdDef.CommandBase + " " + strings.Join(args, " ")
		b.WriteString(m.theme.Box.Render("Command Preview:\n" + m.theme.Accent.Render(fullCmdStr)))
		b.WriteString("\n\n")

		b.WriteString(m.theme.Help.Render("↑/↓ Focus parameter   Enter Run Command   ctrl+g Ask AI   esc Back"))

	case cmdHubConfirmDestructive:
		b.WriteString(m.theme.Warning.Render("▲ CAUTION: DESTRUCTIVE ACTION ▲"))
		b.WriteString("\n\n")
		args := commands.BuildArgs(m.selectedCmd, m.paramValues)
		fullCmdStr := m.selectedCmd.CommandBase + " " + strings.Join(args, " ")
		warningMsg := fmt.Sprintf(
			"You are about to execute a destructive command:\n\n  %s\n\nThis could overwrite files, clean untracked data, or delete remote settings. Do you want to proceed?",
			m.theme.Error.Render(fullCmdStr),
		)
		b.WriteString(m.theme.Box.Render(warningMsg))
		b.WriteString("\n\n")
		b.WriteString(m.theme.Help.Render("y Yes, execute   n No, abort"))

	case cmdHubExecuting:
		b.WriteString(m.theme.Header.Render("Running command..."))
		b.WriteString("\n\n")
		args := commands.BuildArgs(m.selectedCmd, m.paramValues)
		fullCmdStr := m.selectedCmd.CommandBase + " " + strings.Join(args, " ")
		
		loadingContent := fmt.Sprintf(
			"  %s  %s\n\n  %s",
			m.spinner.View(),
			m.theme.Text.Bold(true).Render("Executing shell command..."),
			m.theme.Accent.Render(fullCmdStr),
		)
		b.WriteString(m.theme.Box.Render(loadingContent))
		b.WriteString("\n\n")
		b.WriteString(m.theme.Help.Render("Please wait..."))

	case cmdHubExecFinished:
		b.WriteString(m.theme.Header.Render("Execution Result"))
		b.WriteString("\n\n")

		statusBadge := m.theme.Success.Render("[ SUCCESS ]")
		if !m.runResult.Success {
			statusBadge = m.theme.Error.Render("[ FAILED ]")
		}

		b.WriteString("Command: ")
		b.WriteString(m.theme.Accent.Render(m.runResult.FullCmd))
		b.WriteString("  ")
		b.WriteString(statusBadge)
		b.WriteString("\n\n")

		// Scrollable Viewport
		viewportHeight := m.height - 11
		if viewportHeight < 3 {
			viewportHeight = 3
		}

		lines := strings.Split(m.runResult.Output, "\n")
		end := m.scrollOffset + viewportHeight
		if end > len(lines) {
			end = len(lines)
		}

		var viewportContent strings.Builder
		for i := m.scrollOffset; i < end; i++ {
			viewportContent.WriteString(lines[i] + "\n")
		}

		b.WriteString(m.theme.Box.Height(viewportHeight + 2).Render(viewportContent.String()))
		b.WriteString("\n\n")

		b.WriteString(m.theme.Help.Render("↑/↓ Scroll output   r Rerun   ctrl+g Explain with AI   esc Back"))

	case cmdHubExplainAI:
		b.WriteString(m.theme.Header.Render("AI Explanation"))
		b.WriteString("\n\n")

		if m.aiLoading {
			b.WriteString(m.spinner.View() + " Communicating with AI provider...")
		} else {
			viewportHeight := m.height - 10
			if viewportHeight < 3 {
				viewportHeight = 3
			}

			lines := strings.Split(m.aiExplainText, "\n")
			end := m.scrollOffset + viewportHeight
			if end > len(lines) {
				end = len(lines)
			}

			var viewportContent strings.Builder
			for i := m.scrollOffset; i < end; i++ {
				viewportContent.WriteString(lines[i] + "\n")
			}

			b.WriteString(m.theme.Box.Height(viewportHeight + 2).Render(viewportContent.String()))
		}
		b.WriteString("\n\n")
		b.WriteString(m.theme.Help.Render("↑/↓ Scroll   esc Back"))
	}

	return lipgloss.NewStyle().Width(m.width).Height(m.height).Render(b.String())
}

func getViewportRange(selected, total, visible int) (int, int) {
	start := 0
	if selected >= visible {
		start = selected - visible + 1
	}
	end := start + visible
	if end > total {
		end = total
	}
	if end-start < visible && start > 0 {
		start = end - visible
		if start < 0 {
			start = 0
		}
	}
	return start, end
}
