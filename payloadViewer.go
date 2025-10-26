package main

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/google/gopacket"
)

type payloadViewer struct {
	currentData string
	packets     chan gopacket.Packet
}

func (m *payloadViewer) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg.(type) {
	case packetMsg:
		m.currentData = msg.(packetMsg).Metadata().Timestamp.Format(time.RFC3339)
	default:
		m.currentData = "no packets yet"
	}
	//m.currentData = "e"
	return m, nil
}

func (m *payloadViewer) Init() tea.Cmd {
	m.currentData = ""
	m.packets = make(chan gopacket.Packet)
	return nil
}

func (m *payloadViewer) View() string {

	return m.currentData
}
