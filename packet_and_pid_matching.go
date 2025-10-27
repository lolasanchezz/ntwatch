package main

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

type socketsInfo struct {
	sockets socketMap
	timeout time.Time
}

type pktFromWire struct {
	IP   string
	Port uint16
}
type RefreshSocketTableMsg struct{}
type SocketTableRefreshedMsg struct{ socks *socketMap }
type matchedPkt struct {
	socketKey  pktFromWire
	socketInfo *socketsDef
}

func (info *socketsInfo) manageProcesses(msg tea.Msg) tea.Cmd {

	switch msg.(type) {
	case pktFromWire:
		return func() tea.Msg {
			return info.matchPktToProcess(msg.(pktFromWire))
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

func (info *socketsInfo) matchPktToProcess(pkt pktFromWire) tea.Cmd {
	for i, sock := range info.sockets {
		if (i.DestIP == pkt.IP) && (i.Port == int32(pkt.Port)) {
			return func() tea.Msg {
				return matchedPkt{socketKey: pkt, socketInfo: sock}
			}
		}
	}
	return func() tea.Msg {
		return RefreshSocketTableMsg{}
	}
}

func refreshSockets() tea.Cmd {
	return func() tea.Msg {
		return SocketTableRefreshedMsg{socks: getCStruct()}
	}
}
