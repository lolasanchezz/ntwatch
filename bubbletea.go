package main

import "C"
import (
	"strconv"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/google/gopacket/pcap"
)

type TickMsg time.Time

func doTick() tea.Cmd {
	return tea.Tick(time.Second/2, func(t time.Time) tea.Msg {
		return TickMsg(t)
	})
}

type model struct {
	handle          *pcap.Handle
	unmatchedPacket []PacketInfo
	display         string
	timer           int
	socketTable     *socketMap
	matchedPackets  []processAndPacket
}

func initialModel() model {
	return model{}
}

func (m model) Init() tea.Cmd {
	return tea.Batch((func() tea.Msg { return wireInit() }), doTick())
}

func (m model) View() string {

	return strconv.Itoa(m.timer) + "\n" + m.display + "\n"
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
		//	m.display = msg.data.destIP //for testing
		cmds = append(cmds, m.matchPacketsCmd(msg))
	case TickMsg:
		m.timer++
		cmds = append(cmds, doTick(), getCStructCmd())
		//refresh socket table
	case socketMapMsg:
		m.socketTable = msg.data
		/* testing
		for _, val := range *m.socketTable {
			m.display = val.ProcessName
			return m, tea.Batch(cmds...)
		}
		*/
		//todo - look at unmatched packets
	case matchedPacketMsg:
		if msg.data.process == (socketKey{}) { //unmatched
			m.unmatchedPacket = append(m.unmatchedPacket, msg.data.packet)
			cmds = append(cmds, getCStructCmd())
			m.display = "unmatched"
		} else {
			m.matchedPackets = append(m.matchedPackets, msg.data)
			m.display = msg.data.packet.destIP

		}

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		}
	}
	return m, tea.Batch(cmds...)
}
