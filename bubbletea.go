package main

import (
	"io"
	"log"

	"github.com/google/gopacket"
	"github.com/google/gopacket/pcap"

	tea "github.com/charmbracelet/bubbletea"
)

type model struct {
	payloadViewer *payloadViewer
	packetChannel chan gopacket.Packet
	packetSource  *gopacket.PacketSource
}

type packetMsg gopacket.Packet

func initialModel() model {
	m := model{
		packetChannel: make(chan gopacket.Packet),
		payloadViewer: &payloadViewer{
			packets: make(chan gopacket.Packet),
		},
	}
	return m
}

func (m model) Init() tea.Cmd {
	handle, err := pcap.OpenLive("en0", 1600, true, pcap.BlockForever)
	if err != nil {
		panic(err)
	}
	m.packetSource = gopacket.NewPacketSource(handle, handle.LinkType())
	return tea.Batch(
		waitForPacket(m.packetChannel),
		readFromWire(m.packetChannel, m.packetSource),
	)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Delegate to payloadViewer's Update method
	var cmds tea.BatchMsg
	//updating payload viewer
	updatedViewer, viewerCmd := m.payloadViewer.Update(msg)
	m.payloadViewer = updatedViewer.(*payloadViewer)
	cmds = append(cmds, viewerCmd)

	switch msg := msg.(type) {
	case packetMsg:
		return m, tea.Batch(append(cmds, waitForPacket(m.packetChannel))...)
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC:
			return m, tea.Quit

		default:
			return m, tea.Batch(cmds...)
		}
	}
	return m, tea.Batch(cmds...)
}

func (m model) View() string {
	return m.payloadViewer.View()
}

func readFromWire(packetChannel chan gopacket.Packet, packetSource *gopacket.PacketSource) tea.Cmd {

	return func() tea.Msg {
		for {
			p, err := packetSource.NextPacket()
			if err == io.EOF {
				return nil //TODO: fix later

			} else if err != nil {
				log.Println("Error:", err)
				return nil
			}

			packetChannel <- p
		}
	}
}

func waitForPacket(packetChannel chan gopacket.Packet) tea.Cmd {
	return func() tea.Msg {
		return packetMsg(<-packetChannel)
	}
}
