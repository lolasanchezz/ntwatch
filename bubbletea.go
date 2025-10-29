package main

import (
	"io"
	"log"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/pcap"

	tea "github.com/charmbracelet/bubbletea"
)

type model struct {
	payloadViewer *payloadViewer
	packetChannel chan gopacket.Packet
	packetSource  *gopacket.PacketSource
	sockets       socketsInfo
}

type entirePacket struct {
	packet  gopacket.Packet
	process socketsDef
}

// recieved message from wire
type packetMsg gopacket.Packet

func initialModel() model {
	m := model{
		payloadViewer: &payloadViewer{},
		sockets:       socketsInfo{sockets: make(socketMap), timeout: time.Now()},
	}

	// Initialize packet source here since Init() receives model by value
	handle, err := pcap.OpenLive("en0", 1600, true, pcap.BlockForever)
	if err != nil {
		panic(err)
	}
	m.packetSource = gopacket.NewPacketSource(handle, handle.LinkType())

	m.packetChannel = make(chan gopacket.Packet)
	return m
}

func (m model) Init() tea.Cmd {
	return tea.Batch(
		waitForPacket(m.packetChannel),
		readFromWire(m.packetChannel, m.packetSource),
		refreshSockets(),
		m.payloadViewerInit(),
	)
}

func (m model) View() string {
	return m.payloadViewer.View()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	m, viewerCmd := m.payloadViewerUpdate(msg)

	var cmds []tea.Cmd
	if viewerCmd != nil {
		cmds = append(cmds, viewerCmd)
	}

	switch msg := msg.(type) {
	case packetMsg:
		cmds = append(cmds, waitForPacket(m.packetChannel), m.sockets.manageProcesses(msg))
	case PacketInfo, RefreshSocketTableMsg, SocketTableRefreshedMsg:
		cmds = append(cmds, []tea.Cmd{m.sockets.manageProcesses(msg)}...)
	case tea.KeyMsg:
		if msg.Type == tea.KeyCtrlC {
			return m, tea.Quit
		}
	}

	return m, tea.Batch(cmds...)
}

func readFromWire(packetChannel chan gopacket.Packet, packetSource *gopacket.PacketSource) tea.Cmd {
	return func() tea.Msg {
		go func() {
			for {
				p, err := packetSource.NextPacket()
				if err == io.EOF {
					return
				} else if err != nil {
					log.Println("Error:", err)
					continue
				}
				packetChannel <- p
			}
		}()
		return nil
	}
}

func waitForPacket(packetChannel chan gopacket.Packet) tea.Cmd {
	return func() tea.Msg {
		return packetMsg(<-packetChannel)
	}
}
