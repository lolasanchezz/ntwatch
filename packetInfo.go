package main

import (
	"bytes"
	"fmt"
	"net"
	"time"

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

// the interface packets are captured on
const captureInterface = "en0"

type wireInitMsg struct {
	handle *pcap.Handle
	mac    net.HardwareAddr // en0's own address, to tell sent packets from received ones
}

func wireInit() Msg {
	// first connect to en0
	inactive, err := pcap.NewInactiveHandle(captureInterface)
	if err != nil {
		return errMsg{fmt.Errorf("opening %s: %w", captureInterface, err)}
	}
	defer inactive.CleanUp()
	inactive.SetSnapLen(1600)
	// wifi on macOS can't do promiscuous mode (OpenLive silently skipped it), and we
	// only care about this machine's own traffic anyway
	inactive.SetPromisc(false)
	inactive.SetTimeout(pcap.BlockForever)
	// hand over each packet as soon as it arrives - otherwise macOS holds them
	// until its buffer fills, which looks like a freeze when traffic is light
	inactive.SetImmediateMode(true)

	handle, err := inactive.Activate()
	if err != nil {
		return errMsg{fmt.Errorf("opening %s: %w", captureInterface, err)}
	}

	var mac net.HardwareAddr
	if iface, err := net.InterfaceByName(captureInterface); err == nil {
		mac = iface.HardwareAddr
	}
	return wireInitMsg{handle: handle, mac: mac}

}

type WireDataMsg struct{ data gopacket.Packet }

func sendPacketCmd(handle *pcap.Handle) Cmd {
	return func() Msg {
		data, ci, err := handle.ReadPacketData()
		if err != nil {
			return errMsg{fmt.Errorf("reading packet: %w", err)}
		}

		packet := gopacket.NewPacket(data, layers.LayerTypeEthernet, gopacket.Default)
		packet.Metadata().CaptureInfo = ci // real capture timestamp + length
		return WireDataMsg{data: packet}
	}
}

type packetInfoMsg struct {
	data    PacketInfo
	retried bool // second try at matching it
}

func readPacketCmd(packet gopacket.Packet, mac net.HardwareAddr) Cmd {
	return func() Msg {
		return packetInfoMsg{data: getIPs(packet, mac)}
	}
}

func getIPs(p gopacket.Packet, mac net.HardwareAddr) PacketInfo {

	var pastIPs [2]gopacket.Endpoint
	var src gopacket.Endpoint
	var dest gopacket.Endpoint
	var packetInfo PacketInfo

	if p.ApplicationLayer() != nil {
		packetInfo.appData = p.ApplicationLayer().Payload()
	}
	if eth, ok := p.LinkLayer().(*layers.Ethernet); ok {
		packetInfo.outgoing = bytes.Equal(eth.SrcMAC, mac)
		// the low bit of the first byte marks broadcast + multicast addresses
		packetInfo.broadcast = len(eth.DstMAC) > 0 && eth.DstMAC[0]&1 == 1
	}
	packetInfo.Timestamp = p.Metadata().Timestamp
	packetInfo.packetSize = p.Metadata().Length
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
				packetInfo.destIP = net.IP(arpLayer.DstProtAddress).String()
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

		// the raw payload - gopacket decodes 443 as tls and hides it otherwise
		payload := p.TransportLayer().LayerPayload()
		packetInfo.readable = readableText(payload, packetInfo.sourcePort, packetInfo.destPort)
		if name := clientHelloSNI(payload); name != "" {
			packetInfo.hostnames = append(packetInfo.hostnames, hostname{ip: packetInfo.destIP, name: name})
		}
	}
	packetInfo.hostnames = append(packetInfo.hostnames, dnsHostnames(p)...)

	return packetInfo

}
