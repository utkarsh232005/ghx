package app

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/utkarshpatrikar/ghx/internal/ai"
	"github.com/utkarshpatrikar/ghx/internal/gh"
	"github.com/utkarshpatrikar/ghx/internal/git"
)

type Model struct {
	state State
}

type statusLoadedMsg struct {
	files []git.FileStatus
	info  git.RepoInfo
	err   error
}

type configLoadedMsg struct {
	config ai.Config
	err    error
}

type diffLoadedMsg struct {
	text string
	err  error
}

type commandFinishedMsg struct {
	output string
	err    error
}

type issuesLoadedMsg struct {
	issues []string
	err    error
}

func New() Model {
	return Model{
		state: State{
			Screen: ScreenHome,
			Menu:   defaultMenu(),
			Config: ai.DefaultConfig(),
		},
	}
}

func (m Model) Init() tea.Cmd {
	return loadConfig
}

func loadConfig() tea.Msg {
	config, err := ai.LoadConfig()
	return configLoadedMsg{config: config, err: err}
}

func loadStatus() tea.Msg {
	files, err := git.Status(".")
	if err != nil {
		return statusLoadedMsg{err: err}
	}
	info, err := git.Info(".")
	return statusLoadedMsg{files: files, info: info, err: err}
}

func loadDiff(paths []string) tea.Cmd {
	return func() tea.Msg {
		text, err := git.Diff(paths, false)
		return diffLoadedMsg{text: text, err: err}
	}
}

func runCommit(paths []string, message string) tea.Cmd {
	return func() tea.Msg {
		result, err := git.Commit(paths, message)
		return commandFinishedMsg{output: result.Output, err: err}
	}
}

func runPush() tea.Msg {
	result, err := git.Push()
	return commandFinishedMsg{output: result.Output, err: err}
}

func runCreatePR(title string, body string) tea.Cmd {
	return func() tea.Msg {
		output, err := gh.CreatePR(title, body)
		return commandFinishedMsg{output: output, err: err}
	}
}

func runCheckoutBranch(name string) tea.Cmd {
	return func() tea.Msg {
		result, err := git.CheckoutBranch(name)
		return commandFinishedMsg{output: result.Output, err: err}
	}
}

func loadIssues() tea.Msg {
	issues, err := gh.IssueList()
	return issuesLoadedMsg{issues: issues, err: err}
}
