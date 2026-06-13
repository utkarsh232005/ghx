package screens

import "github.com/KDM-cli/ghx/internal/app"

type NavigateMsg struct {
	Screen app.Screen
}

func Navigate(screen app.Screen) NavigateMsg {
	return NavigateMsg{Screen: screen}
}
