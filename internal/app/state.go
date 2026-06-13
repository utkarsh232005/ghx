package app

import (
	"github.com/utkarshpatrikar/ghx/internal/ai"
	"github.com/utkarshpatrikar/ghx/internal/components"
	"github.com/utkarshpatrikar/ghx/internal/git"
)

type Screen int

const (
	ScreenHome Screen = iota
	ScreenStatus
	ScreenCommit
	ScreenDiff
	ScreenPush
	ScreenPR
	ScreenIssues
	ScreenRepos
	ScreenAI
	ScreenHistory
	ScreenSettings
)

type State struct {
	Screen        Screen
	Width         int
	Height        int
	Menu          components.Menu
	StatusFiles   []git.FileStatus
	RepoInfo      git.RepoInfo
	Config        ai.Config
	Message       string
	Loading       bool
	Err           error
	FileCursor    int
	SelectedFiles map[string]bool
	CommitMessage string
	CommitFocus   int
	DiffText      string
	DiffScroll    int
	PushConfirm   bool
	Output        string
	PRTitle       string
	PRBody        string
	PRFocus           int
	BranchInputActive bool
	NewBranchName     string
	Issues            []string
	AIInput           string
	AIResponse        string
}

func defaultMenu() components.Menu {
	return components.NewMenu([]components.MenuItem{
		{Title: "Status", Description: "View git status", Target: "status"},
		{Title: "Commit", Description: "Stage & commit files (AI assist)", Target: "commit"},
		{Title: "Diff", Description: "View staged changes", Target: "diff"},
		{Title: "Push", Description: "Push to remote", Target: "push"},
		{Title: "PR", Description: "Create pull request (AI assist)", Target: "pr"},
		{Title: "Issues", Description: "Manage issues", Target: "issues"},
		{Title: "Repos", Description: "Repo navigation and metadata", Target: "repos"},
		{Title: "AI Chat", Description: "Ask AI about codebase", Target: "ai"},
		{Title: "History", Description: "View recent commands", Target: "history"},
		{Title: "Settings", Description: "Configure providers & preferences", Target: "settings"},
	})
}
