package main

import "github.com/google/gopacket/layers"

type processAndPacket struct {
	process socketKey
	packet  PacketInfo
	system  bool // unmatched, but nothing could have matched it - broadcast/multicast, or no ports (ARP etc)
}

type matchedPacketMsg struct {
	data    processAndPacket
	retried bool
}

func protoOf(packet PacketInfo) (int32, bool) {
	switch packet.packetProtocol {
	case layers.LayerTypeTCP:
		return TCP, true
	case layers.LayerTypeUDP:
		return UDP, true
	}
	return 0, false
}

func sameProto(sock socketKey, packet PacketInfo) bool {
	proto, ok := protoOf(packet)
	return ok && sock.Proto == proto
}

// tries network statistics flows first (exact, and they see what the socket scan
// can't), then the socket table, then unconnected/listening sockets by local port
func matchPackets(flows flowIndex, socketTable socketMap, packet packetInfoMsg) processAndPacket {
	// broadcast/multicast from another machine reaches every listener on that port,
	// so there's no one process to give it to. (one we sent still has a sender)
	if packet.data.broadcast && !packet.data.outgoing {
		return processAndPacket{packet: packet.data, system: true}
	}

	// the port on this machine's end of the packet
	localPort := packet.data.destPort
	if packet.data.outgoing {
		localPort = packet.data.sourcePort
	}

	proto, hasPorts := protoOf(packet.data)
	if hasPorts {
		sent := flowKey{proto: proto, localPort: packet.data.sourcePort, remoteIP: packet.data.destIP, remotePort: packet.data.destPort}
		received := flowKey{proto: proto, localPort: packet.data.destPort, remoteIP: packet.data.sourceIP, remotePort: packet.data.sourcePort}
		if f, ok := flows[sent]; ok {
			return processAndPacket{process: f.sock, packet: packet.data}
		}
		if f, ok := flows[received]; ok {
			return processAndPacket{process: f.sock, packet: packet.data}
		}
	}

	// unconnected udp / listening tcp sockets (dest "0") have no remote end to compare
	// against, so they only get a packet if no connected socket fits it
	var fallback socketKey
	for _, val := range socketTable {
		if !sameProto(val, packet.data) {
			continue
		}
		if val.DestPort == "0" {
			if val.SrcPort == localPort && fallback == (socketKey{}) {
				fallback = val
			}
			continue
		}

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
	if hasPorts {
		if f, ok := flows[flowKey{proto: proto, localPort: localPort, remoteIP: "0", remotePort: "0"}]; ok {
			return processAndPacket{process: f.sock, packet: packet.data}
		}
	}
	if fallback != (socketKey{}) {
		return processAndPacket{process: fallback, packet: packet.data}
	}
	return processAndPacket{
		packet: packet.data,
		system: packet.data.broadcast || packet.data.sourcePort == "",
	}
}

// the flow index and socket table are never modified after they're built, so
// they're safe to hand to a cmd
func matchPacketsCmd(flows flowIndex, socketTable socketMap, packet packetInfoMsg) Cmd {
	return func() Msg {
		return matchedPacketMsg{data: matchPackets(flows, socketTable, packet), retried: packet.retried}
	}
}
