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
}

func initialModel() model {

	return model{}
}

func (m model) Init() tea.Cmd {
	go func() {
		readFromWire(m.payloadViewer.packets)
	}()
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Delegate to payloadViewer's Update method
	var cmd tea.Cmd
	updatedViewer, viewerCmd := m.payloadViewer.Update(msg)
	m.payloadViewer = updatedViewer.(*payloadViewer)
	cmd = viewerCmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC:
			return m, tea.Quit
		}
	}
	return m, cmd
}

func (m model) View() string {
	return m.payloadViewer.View()
}

func readFromWire(ch chan<- gopacket.Packet) {
	handle, err := pcap.OpenLive("en0", 1600, true, pcap.BlockForever)
	if err != nil {
		panic(err)
	}

	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())

	for {
		/*
			packet := getIPs(*packetSource)
			if packet.packetProtocol != layers.LayerTypeTCP {
				//fmt.Print(packet.packetProtocol, "\n")
			}
			if packet.appData != nil {
				fmt.Print(lookThroughBody(packet.appData))
			}
		*/
		p, err := packetSource.NextPacket()
		if err == io.EOF {
			return //TODO: fix later

		} else if err != nil {
			log.Println("Error:", err)
			return
		}
		ch <- p

	}

}
