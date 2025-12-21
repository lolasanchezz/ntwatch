package main

import tea "github.com/charmbracelet/bubbletea"

type processAndPacket struct {
	process socketKey
	packet  PacketInfo
}

type matchedPacketMsg struct {
	data processAndPacket
}

func (m model) matchPackets(packet packetInfoMsg) processAndPacket {
	if m.socketTable == nil {
		return processAndPacket{packet: packet.data}
	}

	for _, val := range *m.socketTable {
		// Outgoing: packet src port = local port, packet dest = remote
		outgoing := (packet.data.destIP == val.DestIP) &&
			(packet.data.sourcePort == val.SrcPort) &&
			(packet.data.destPort == val.DestPort)
		// Incoming: packet dest port = local port, packet src = remote
		incoming := (packet.data.sourceIP == val.DestIP) &&
			(packet.data.destPort == val.SrcPort) &&
			(packet.data.sourcePort == val.DestPort)

		if outgoing || incoming {
			return processAndPacket{
				process: val,
				packet:  packet.data,
			}
		}
	}
	return processAndPacket{packet: packet.data}
}

func (m model) matchPacketsCmd(packet packetInfoMsg) tea.Cmd {
	return func() tea.Msg {
		return matchedPacketMsg{data: m.matchPackets(packet)}
	}
}
