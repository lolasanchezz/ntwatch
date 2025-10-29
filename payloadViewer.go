package main

import (
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
)

type payloadViewer struct {
	currentData string
	viewport    viewport.Model
}

func (m model) payloadViewerUpdate(msg tea.Msg) (model, tea.Cmd) {
	if m.payloadViewer.viewport.Width == 0 && m.payloadViewer.viewport.Height == 0 {
		// Initialize viewport if not already initialized
		m.payloadViewer.viewport = viewport.New(80, 24)
	}
	m.payloadViewer.viewport.SetContent(m.payloadViewer.currentData)
	/*
		switch msg.(type) {
		case packetMsg:
			m.payloadViewer.currentData = msg.(packetMsg).NetworkLayer().NetworkFlow().Dst().String()
		default:
			m.payloadViewer.currentData = "no packets yet"
		}
		//m.currentData = "e"
	*/
	return m, nil
}

func (m model) payloadViewerInit() tea.Cmd {
	m.payloadViewer.currentData = ""
	return nil
}

func (m *payloadViewer) View() string {
	if m.viewport.Width == 0 && m.viewport.Height == 0 {
		// Return default content if viewport not initialized
		if m.currentData == "" {
			return "No packets yet..."
		}
		return m.currentData
	}
	return m.viewport.View()
}
