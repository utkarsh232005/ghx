package app

import (
	"github.com/charmbracelet/bubbletea"
	"github.com/KDM-cli/ghx/internal/ai"
	"github.com/KDM-cli/ghx/internal/db"
	"github.com/KDM-cli/ghx/internal/screens"
	"github.com/KDM-cli/ghx/styles"
)

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
)

type Model struct {
	currentScreen Screen
	screens       map[Screen]tea.Model
	db            *db.DB
	aiManager     *ai.Manager
	theme         *styles.Theme
	width         int
	height        int
	err           error
}

func New() Model {
	theme := styles.NewTheme()

	database, err := db.New()
	if err != nil {
		return Model{
			theme: theme,
			err:   err,
		}
	}

	aiManager := ai.NewManager(database)

	screenModels := map[Screen]tea.Model{
		ScreenHome:     screens.NewHomeModel(theme, aiManager),
		ScreenStatus:   screens.NewStatusModel(theme, database),
		ScreenCommit:   screens.NewCommitModel(theme, database, aiManager),
		ScreenPush:     screens.NewPushModel(theme, database),
		ScreenPR:       screens.NewPRModel(theme, aiManager),
		ScreenIssues:   screens.NewIssuesModel(theme),
		ScreenRepos:    screens.NewReposModel(theme),
		ScreenAIChat:   screens.NewAIChatModel(theme, aiManager),
		ScreenSettings: screens.NewSettingsModel(theme, aiManager, database),
		ScreenHelp:     screens.NewHelpModel(theme),
	}

	return Model{
		currentScreen: ScreenHome,
		screens:       screenModels,
		db:            database,
		aiManager:     aiManager,
		theme:         theme,
	}
}

func (m Model) Init() tea.Cmd {
	return m.screens[m.currentScreen].Init()
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "q":
			if m.currentScreen == ScreenHome {
				return m, tea.Quit
			}
		case "esc":
			if m.currentScreen != ScreenHome {
				m.currentScreen = ScreenHome
				return m, m.screens[ScreenHome].Init()
			}
		case "?":
			if m.currentScreen != ScreenHelp {
				prevScreen := m.currentScreen
				m.currentScreen = ScreenHelp
				return m, func() tea.Msg {
					return screens.HelpInitMsg{PreviousScreen: prevScreen}
				}
			}
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		for screenName, screen := range m.screens {
			updated, _ := screen.Update(msg)
			m.screens[screenName] = updated
		}
		return m, nil

	case screens.NavigateMsg:
		if targetScreen, ok := m.screens[msg.Screen]; ok {
			m.currentScreen = msg.Screen
			return m, targetScreen.Init()
		}
	}

	updatedScreen, cmd := m.screens[m.currentScreen].Update(msg)
	m.screens[m.currentScreen] = updatedScreen
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	if m.err != nil {
		return m.theme.Error.Render("Error: " + m.err.Error())
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
