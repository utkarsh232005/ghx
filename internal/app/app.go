package app

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/KDM-cli/ghx/internal/ai"
	"github.com/KDM-cli/ghx/internal/db"
	"github.com/KDM-cli/ghx/internal/screens"
	"github.com/KDM-cli/ghx/styles"
)

type Model struct {
	currentScreen screens.Screen
	screens       map[screens.Screen]tea.Model
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

	screenModels := map[screens.Screen]tea.Model{
		screens.ScreenHome:     screens.NewHomeModel(theme, aiManager),
		screens.ScreenStatus:   screens.NewStatusModel(theme, database),
		screens.ScreenCommit:   screens.NewCommitModel(theme, database, aiManager),
		screens.ScreenPush:     screens.NewPushModel(theme, database),
		screens.ScreenPR:       screens.NewPRModel(theme, aiManager),
		screens.ScreenIssues:   screens.NewIssuesModel(theme),
		screens.ScreenRepos:    screens.NewReposModel(theme),
		screens.ScreenAIChat:   screens.NewAIChatModel(theme, aiManager),
		screens.ScreenSettings: screens.NewSettingsModel(theme, aiManager, database),
		screens.ScreenHelp:     screens.NewHelpModel(theme),
	}

	return Model{
		currentScreen: screens.ScreenHome,
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
			if m.currentScreen == screens.ScreenHome {
				return m, tea.Quit
			}
		case "esc":
			if m.currentScreen != screens.ScreenHome {
				m.currentScreen = screens.ScreenHome
				return m, m.screens[screens.ScreenHome].Init()
			}
		case "?":
			if m.currentScreen != screens.ScreenHelp {
				prevScreen := m.currentScreen
				m.currentScreen = screens.ScreenHelp
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
	if current, ok := m.screens[m.currentScreen]; ok {
		return current.View()
	}
	return "Screen not found"
}
