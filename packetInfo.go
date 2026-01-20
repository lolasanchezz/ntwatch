package main

import (
	"net"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
)

type packetProtocol string

/*
const (

	Arp  packetProtocol = "Arp"
	IPv4 packetProtocol = "IPv4"
	IPv6 packetProtocol = "IPv6"
	SSDP packetProtocol = "SSDP"
	DNS  packetProtocol = "DNS"
	TLS  packetProtocol = "TLS"
	TCP  packetProtocol = "TCP"

)
*/

type wireInitMsg struct{ handle *pcap.Handle }

func wireInit() wireInitMsg {
	// first connect to en0
	handle, err := pcap.OpenLive("en0", 1600, true, pcap.BlockForever)
	if err != nil {
		panic(err)
	}

	return wireInitMsg{handle: handle}

}

type WireDataMsg struct{ data gopacket.Packet }

func (m model) sendPacketCmd() tea.Cmd {
	return func() tea.Msg {
		packet, _, err := m.handle.ReadPacketData()
		if err != nil {
			panic(err)
		}

		return WireDataMsg{
			data: gopacket.NewPacket(packet, layers.LayerTypeEthernet, gopacket.Default),
		}
	}
}

type packetInfoMsg struct{ data PacketInfo }

func readPacketCmd(packet gopacket.Packet) tea.Cmd {
	return func() tea.Msg {
		return packetInfoMsg{data: getIPs(packet)}
	}
}

func getIPs(p gopacket.Packet) PacketInfo {

	var pastIPs [2]gopacket.Endpoint
	var src gopacket.Endpoint
	var dest gopacket.Endpoint
	var packetInfo PacketInfo

	if p.ApplicationLayer() != nil {
		packetInfo.appData = p.ApplicationLayer().Payload()
	}
	packetInfo.Timestamp = p.Metadata().Timestamp
	// Debug: check if timestamp is zero
	if packetInfo.Timestamp.IsZero() {
		packetInfo.Timestamp = time.Now()
	}
	netLayer := p.NetworkLayer()
	if netLayer == nil {
		for _, layer := range p.Layers() {
			if layer.LayerType() == layers.LayerTypeARP {
				packetInfo.packetProtocol = layers.LayerTypeARP
				arpLayer := layer.(*layers.ARP)
				packetInfo.sourceIP = net.HardwareAddr(arpLayer.SourceHwAddress).String()
				packetInfo.destIP = net.HardwareAddr(arpLayer.DstProtAddress).String()
				//fmt.Print("arp packet sent by ", packetInfo.sourceIP, "looking for", packetInfo.destIP, "\n")
			}
		}
		return packetInfo
	}

	src, dest = netLayer.NetworkFlow().Endpoints()
	packetInfo.destIP = dest.String()
	packetInfo.sourceIP = src.String()

	if !((src == pastIPs[0]) && (dest == pastIPs[1])) && !((src == pastIPs[1]) && (dest == pastIPs[0])) {
		//means that this is a new flow
		//	fmt.Print("\n")
		//		fmt.Print(src, dest, "\n")
		pastIPs[0] = src
		pastIPs[1] = dest
	}

	if p.TransportLayer() != nil {
		packetInfo.packetProtocol = p.TransportLayer().LayerType()
		//	fmt.Print("ports ")
		//	fmt.Print(p.TransportLayer().TransportFlow().Endpoints())
		//	fmt.Print(" \n")
		packetInfo.sourcePort = p.TransportLayer().TransportFlow().Src().String()
		packetInfo.destPort = p.TransportLayer().TransportFlow().Dst().String()
	}

	return packetInfo

}
