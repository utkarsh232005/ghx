package screens

type Screen string

const (
	ScreenHome     Screen = "home"
	ScreenStatus   Screen = "status"
	ScreenCommit   Screen = "commit"
	ScreenPush     Screen = "push"
	ScreenPull     Screen = "pull"
	ScreenBranch   Screen = "branch"
	ScreenPR       Screen = "pr"
	ScreenIssues   Screen = "issues"
	ScreenRepos    Screen = "repos"
	ScreenAIChat   Screen = "ai_chat"
	ScreenSettings Screen = "settings"
	ScreenHelp     Screen = "help"
	ScreenCmdHub   Screen = "cmd_hub"
)

type NavigateMsg struct {
	Screen Screen
}

func Navigate(screen Screen) NavigateMsg {
	return NavigateMsg{Screen: screen}
}
