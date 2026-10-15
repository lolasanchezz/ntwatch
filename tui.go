// bubble tea viewing - just draws whatever snapshot the engine last sent
package main

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
)

var defaultHeaders = []string{
	"Source IP", "Dest IP", "Src Port", "Dest Port", "Connection type", "Time",
}

type snapshotMsg Snapshot

type tuiModel struct {
	engine *Engine
	snap   Snapshot
}

func runTUI() error {
	engine := NewEngine()
	p := tea.NewProgram(tuiModel{engine: engine}, tea.WithMouseAllMotion(), tea.WithAltScreen())

	go func() {
		engine.Run(initialModel(func(s Snapshot) { p.Send(snapshotMsg(s)) }))
		p.Quit()
	}()

	_, err := p.Run()
	engine.Stop()
	return err
}

func (t tuiModel) Init() tea.Cmd {
	return nil
}

func (t tuiModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case snapshotMsg:
		t.snap = Snapshot(msg)
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return t, tea.Quit

		case "q":
			return t, tea.Quit
		}
	}
	return t, nil
}

func (t tuiModel) View() string {
	if t.snap.Err != "" {
		return "error: " + t.snap.Err + "\n"
	}
	// Only skip if no sockets exist yet
	if len(t.snap.RecentSockets) == 0 {
		return ""
	}
	//can change later
	tableNum := min(len(t.snap.RecentSockets), 3)

	render := ""
	recentSocketsTable := table.New().Headers("Name", "Pid")
	for _, sock := range t.snap.RecentSockets {
		recentSocketsTable.Row(sock.ProcessName, sock.Pid)
	}
	for _, sock := range t.snap.RecentSockets[:tableNum] {
		table := table.New().Headers(defaultHeaders...)
		for _, p := range sock.Packets {
			table.Row(
				p.SourceIP,
				p.DestIP,
				p.SourcePort,
				p.DestPort,
				p.Protocol,
				time.UnixMilli(p.Time).Format("15:04:05"),
			)
		}
		render = lipgloss.JoinVertical(lipgloss.Center, render, (sock.ProcessName + " " + sock.Pid), table.Render(), "\n")
	}
	render = lipgloss.JoinHorizontal(lipgloss.Top, render, recentSocketsTable.Render())
	//for debugging
	render = lipgloss.JoinVertical(lipgloss.Top, (time.Now().Format("15:04:05.000")), render)
	return render
}
