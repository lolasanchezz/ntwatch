package main

/*
import (
	"fmt"
	"io"
	"log"
	"net"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
)

type packetProtocol string


const (

	Arp  packetProtocol = "Arp"
	IPv4 packetProtocol = "IPv4"
	IPv6 packetProtocol = "IPv6"
	SSDP packetProtocol = "SSDP"
	DNS  packetProtocol = "DNS"
	TLS  packetProtocol = "TLS"
	TCP  packetProtocol = "TCP"

)


func givePackets() PacketInfo {
	// first connect to en0
	handle, err := pcap.OpenLive("en0", 1600, true, pcap.BlockForever)
	if err != nil {
		panic(err)
	}

	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())

	return getIPs(*packetSource)

}

func handlePacketsv1(packet gopacket.PacketSource, ch chan string) {
	fmt.Print("running")

	for p := range packet.Packets() {
		netLayer := p.NetworkLayer()
		if netLayer == nil {
			continue
		}
		src, dest := netLayer.NetworkFlow().Endpoints()
		srcIp := fmt.Sprint(src)
		destIp := fmt.Sprint(dest)
		fmt.Print(srcIp, destIp)
		//ch <- fmt.Sprint(src)
		//ch <- fmt.Sprint(dest)

	}
}

func getIPs(packet gopacket.PacketSource) PacketInfo {

	var pastIPs [2]gopacket.Endpoint
	var src gopacket.Endpoint
	var dest gopacket.Endpoint
	var packetInfo PacketInfo
	//for {

	p, err := packet.NextPacket()
	if err == io.EOF {
		packetInfo.eof = true
		return packetInfo

	} else if err != nil {
		log.Println("Error:", err)
		return packetInfo
	}

	if p.ApplicationLayer() != nil {
		packetInfo.appData = p.ApplicationLayer().Payload()
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
	if !((src == pastIPs[0]) && (dest == pastIPs[1])) && !((src == pastIPs[1]) && (dest == pastIPs[0])) {
		//means that this is a new flow
		//	fmt.Print("\n")
		//		fmt.Print(src, dest, "\n")
		pastIPs[0] = src
		pastIPs[1] = dest
		packetInfo.destIP = dest.String()
		packetInfo.sourceIP = src.String()
		if p.TransportLayer() != nil {
			packetInfo.packetProtocol = p.TransportLayer().LayerType()
			//	fmt.Print("ports ")
			//	fmt.Print(p.TransportLayer().TransportFlow().Endpoints())
			//	fmt.Print(" \n")
			packetInfo.sourcePort = p.TransportLayer().TransportFlow().Src().String()
			packetInfo.destPort = p.TransportLayer().TransportFlow().Dst().String()

		}

	}

	return packetInfo
}
*/
