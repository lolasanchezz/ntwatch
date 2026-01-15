package main

import "C"
import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/google/gopacket/pcap"
)

type TickMsg time.Time

func doTick() tea.Cmd {
	return tea.Tick(time.Second/2, func(t time.Time) tea.Msg {
		return TickMsg(t)
	})
}

type processDesc struct {
	name string
}
type processSocketMap map[processDesc]map[socketKey]PacketInfo

type display struct {
	height int
	width  int
}

type model struct {
	handle            *pcap.Handle
	unmatchedPacket   []PacketInfo
	timer             int
	socketTable       *socketMap
	matchedPackets    []processAndPacket //deprecated - not using anymore
	matchedPacketsTbl map[socketKey][]PacketInfo
	displayTable      *table.Table
	display           display
}

func initialModel() model {
	return model{
		displayTable: table.New().Headers("Name", "Source IP", "Dest IP", "Src Port", "Dest Port"),
	}
}

func (m model) Init() tea.Cmd {

	return tea.Batch((func() tea.Msg { return wireInit() }), doTick())
}

func (m model) View() string {
	if (m.socketTable == nil) || (len(m.matchedPackets) == 0) || (m.display.height == 0) {
		return ""
	}
	rowLen := min(m.display.height-4, len(m.matchedPackets))
	rows := make([][]string, rowLen)
	for i := range rowLen {
		row := (m.matchedPackets)[len(m.matchedPackets)-i-1]
		rows[i] = []string{row.process.ProcessName, row.packet.sourceIP, row.packet.destIP, row.packet.sourcePort, row.packet.destPort}
	}
	m.displayTable.ClearRows()
	m.displayTable.Rows(rows...)
	return m.displayTable.Render()
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

		} else {
			m.matchedPackets = append(m.matchedPackets, msg.data)

		}
	case tea.WindowSizeMsg:
		m.display.height = msg.Height
		m.display.width = msg.Width
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		}
	}
	return m, tea.Batch(cmds...)
}
