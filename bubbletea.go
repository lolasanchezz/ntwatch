package main

import "C"
import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/google/gopacket/pcap"
)

type TickMsg time.Time

func doTick() tea.Cmd {
	return tea.Tick(time.Second*2, func(t time.Time) tea.Msg {
		return TickMsg(t)
	})
}

var defaultHeaders = []string{
	"Source IP", "Dest IP", "Src Port", "Dest Port", "Connection type", "Time",
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
	handle             *pcap.Handle
	unmatchedPacket    []PacketInfo
	timer              int
	tableNum           int
	socketNum          int
	socketTable        *socketMap
	matchedPackets     []processAndPacket //deprecated - not using anymore
	matchedPacketsTbl  map[socketKey][]PacketInfo
	recentSockets      []socketKey
	recentSocketsTable *table.Table
	display            display
}

func initialModel() model {
	tableNum := 5
	socketNum := 10
	return model{
		recentSocketsTable: table.New().Headers("Name", "Pid"),
		matchedPacketsTbl:  make(map[socketKey][]PacketInfo),
		tableNum:           tableNum,
		socketNum:          socketNum,
		recentSockets:      make([]socketKey, 0, tableNum+1), //because it's a queue
	}
}

func (m model) Init() tea.Cmd {

	return tea.Batch((func() tea.Msg { return wireInit() }), doTick())
}

func (m model) View() string {
	// Only skip if no sockets exist yet
	if len(m.recentSockets) == 0 {
		return ""
	}
	//can change later
	tableNum := min(len(m.recentSockets), 3)
	render := ""
	rows := make([][]string, len(m.recentSockets))
	for i, val := range m.recentSockets {
		rows[len(m.recentSockets)-i-1] = []string{val.ProcessName, val.Pid}
	}
	m.recentSocketsTable.ClearRows().Rows(rows...)
	for i := range tableNum {
		table := table.New().Headers(defaultHeaders...)
		socket := m.recentSockets[len(m.recentSockets)-i-1]
		packets := m.matchedPacketsTbl[socket]
		rowCount := min(m.socketNum, len(packets))
		for j := range rowCount {
			table.Rows([]string{
				packets[len(packets)-j-1].sourceIP,
				packets[len(packets)-j-1].destIP,
				packets[len(packets)-j-1].sourcePort,
				packets[len(packets)-j-1].destPort,
				packets[len(packets)-j-1].packetProtocol.String(),
				packets[len(packets)-j-1].Timestamp.Format("15:04:05"),
			})
		}
		render = lipgloss.JoinVertical(lipgloss.Left, render, (socket.ProcessName + " " + socket.Pid), table.Render(), "\n")
	}
	render = lipgloss.JoinHorizontal(lipgloss.Top, render, m.recentSocketsTable.Render())
	return render
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
			// Always append the packet to the table for this socket
			m.matchedPacketsTbl[msg.data.process] = append(m.matchedPacketsTbl[msg.data.process], msg.data.packet)

			// Add to recent sockets if this is a new socket (index == -1)
			if msg.data.index == -1 {
				if len(m.recentSockets) < m.tableNum {
					// Room for more sockets
					m.recentSockets = append(m.recentSockets, msg.data.process)
				} else {
					// Queue is full, remove oldest and add newest
					m.recentSockets = append(m.recentSockets[1:], msg.data.process)
				}
			}
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
