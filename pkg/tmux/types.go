package tmux

// Session represents a tmux session
type Session struct {
	Name  string
	Panes []*Pane
}

// Pane represents a tmux pane
type Pane struct {
	ID      string
	Index   int
	AgentID string
}
