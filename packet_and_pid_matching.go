package main

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/google/gopacket/layers"
)

// struct that stores all sockets and timeout
type socketsInfo struct {
	sockets socketMap
	timeout time.Time
}

// socket table should be refreshed
type RefreshSocketTableMsg struct{}

// socket table IS refreshed
type SocketTableRefreshedMsg struct{ socks *socketMap }

// matched packet
type matchedPkt struct {
	socketKey  packetMsg
	socketInfo *socketsDef
}

func (info *socketsInfo) manageProcesses(msg tea.Msg) tea.Cmd {

	switch msg.(type) {
	case packetMsg:
		return func() tea.Msg {
			return info.matchPktToProcess(msg.(packetMsg))
		}
	case RefreshSocketTableMsg:
		return func() tea.Msg {
			return refreshSockets()
		}
	case SocketTableRefreshedMsg:
		info.timeout = time.Now()
		info.sockets = *msg.(SocketTableRefreshedMsg).socks
		return nil
	}
	return nil
}

func (info *socketsInfo) matchPktToProcess(pkt packetMsg) tea.Cmd {
	//first check type of packet, either cast them to udp or tcp
	var srcPort int32
	var dstPort int32
	var connType int32

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
	for i, sock := range info.sockets {
		if (i.DestIP == destIP) && (i.SrcPort == srcPort) && (i.DestPort == dstPort) && (i.ConnType == connType) {
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

func refreshSockets() tea.Cmd {
	return func() tea.Msg {
		return SocketTableRefreshedMsg{socks: getCStruct()}
	}
}
