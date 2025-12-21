package main

/*

import (
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/google/gopacket/layers"
)
// design
//- match packet to socket as it comes in
	//- when a packet is recieved, attempt to match it to the socket table
//	- if a match is not found, send a message to refresh the socket table
//	- if more than a minute has passed, send a message to refresh the socket table
//	- WHILE a table is being refreshed, do n
//	- if a match is found, insert it into a struct of packet and info to display
//		- then formatted and made pretty by view, but has no viewing/rendering logic




// struct that stores all sockets and timeout
type socketsInfo struct {
	sockets socketMap
	timeout time.Time
}

// socket table should be refreshed
type RefreshSocketTableMsg struct{}

// socket table IS refreshed
type SocketTableRefreshedMsg struct{ socks *socketMap }

// matched packet to process
type matchedPkt struct {
	socketKey  packetMsg
	socketInfo *socketsDef
}


//manage process - a function that's supposed to match incoming packets to already known
//packets, manage the sockets coming in, and signal when to refresh the socket table

func (m *model) manageProcesses(msg tea.Msg) tea.Cmd {

	switch msg.(type) {
	case packetMsg:
		return func() tea.Msg {
			return m.matchPktToProcess(msg.(packetMsg))
		}
	case RefreshSocketTableMsg:
		return func() tea.Msg {
			return m.refreshSockets()
		}
	case SocketTableRefreshedMsg:

		m.timeKeeping.pastTime = time.Now()
		m.sockets = *msg.(SocketTableRefreshedMsg).socks
		return nil

	case tickSecondPassed:
		writeToDebug("second passed \n")
		return func() tea.Msg {
			return m.refreshSockets()
		}
	}
	return nil
}

func (m *model) matchPktToProcess(pkt packetMsg) tea.Cmd {
	//first check type of packet, either cast them to udp or tcp
	var srcPort int32
	var dstPort int32
	var connType int32

	_ = srcPort
	_ = dstPort
	_ = connType
	// Check if packet has a network layer
	netLayer := pkt.NetworkLayer()
	if netLayer == nil {
		return nil
	}

	if tcpLayer := pkt.Layer(layers.LayerTypeTCP); tcpLayer != nil {
		tcp, ok := tcpLayer.(*layers.TCP)
		if ok && tcp != nil {
			srcPort = int32(tcp.SrcPort)
			dstPort = int32(tcp.DstPort)
			connType = 1
		}
	} else if udpLayer := pkt.Layer(layers.LayerTypeUDP); udpLayer != nil {
		udp, ok := udpLayer.(*layers.UDP)
		if ok && udp != nil {
			srcPort = int32(udp.SrcPort)
			dstPort = int32(udp.DstPort)
			connType = 2
		}
	}

	// Get network flow
	netFlow := netLayer.NetworkFlow()
	destIP := netFlow.Dst().String()
	for i, sock := range m.sockets {
		if i.DestIP == destIP { // && (i.SrcPort == srcPort) && (i.DestPort == dstPort) && (i.ConnType == connType) {
			// append debug message to file
			f, err := os.OpenFile("./debug.txt", os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
			if err == nil {
				_, _ = f.WriteString("\n matched")
				_ = f.Close()
			}
			return func() tea.Msg {
				return matchedPkt{socketKey: pkt, socketInfo: sock}
			}
		}
	}
	//if no matched packets are found, refresh the socket table
	return func() tea.Msg {
		return RefreshSocketTableMsg{}
	}
}

func (m *model) refreshSockets() tea.Cmd {
	m.sockets = *getCStruct()

	return func() tea.Msg {
		return SocketTableRefreshedMsg{socks: getCStruct()}
	}
}
*/
