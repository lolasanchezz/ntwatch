package main

import (
	"fmt"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
)

func main() {
	// first connect to en0
	handle, err := pcap.OpenLive("en0", 1600, true, pcap.BlockForever)
	if err != nil {
		panic(err)
	}

	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())

	for {

		packet := getIPs(*packetSource)
		if packet.packetProtocol != layers.LayerTypeTCP {
			//fmt.Print(packet.packetProtocol, "\n")
		}
		if packet.appData != nil {
			fmt.Print(lookThroughBody(packet.appData))
		}
	}

}
