package main

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/google/gopacket"
	"github.com/google/gopacket/pcap"
)

type model struct {
	handle       *pcap.Handle
	unmatchedMsg PacketInfo
	display      string
	socketTable  socketMap
}

// recieved message from wire
type packetMsg gopacket.Packet
type getPacket struct{}

func initialModel() model {
	return model{}
}

func (m model) Init() tea.Cmd {
	return func() tea.Msg { return wireInit() }
}

func (m model) View() string {
	return m.display
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case wireInitMsg:
		m.handle = msg.handle
		cmds = append(cmds, m.sendPacketCmd())
	case WireDataMsg:
		cmds = append(cmds, readPacketCmd(msg.data), m.sendPacketCmd())
	case packetInfoMsg:
		m.display = msg.data.destIP
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		}
	}
	return m, tea.Batch(cmds...)
}
