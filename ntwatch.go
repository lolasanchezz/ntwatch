package main

import (
	"fmt"
	"io"
	"log"

	"github.com/google/gopacket"
	"github.com/google/gopacket/pcap"
)

func main() {
	// first connect to en0
	handle, err := pcap.OpenLive("en0", 1600, true, pcap.BlockForever)
	if err != nil {
		panic(err)
	}

	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())

	//ch := make(chan string)
	/*
		go handlePackets(*packetSource, ch)

		for {
			fmt.Print(<-ch)
		}
	*/
	getIPs(*packetSource)
}

func handlePacketsv1(packet gopacket.PacketSource, ch chan string) {
	fmt.Print("running")

	for p := range packet.Packets() {

		src, dest := p.NetworkLayer().NetworkFlow().Endpoints()
		srcIp := fmt.Sprint(src)
		destIp := fmt.Sprint(dest)
		fmt.Print(srcIp, destIp)
		//ch <- fmt.Sprint(src)
		//ch <- fmt.Sprint(dest)

	}
}

func getIPs(packet gopacket.PacketSource) {
	var pastIPs [2]gopacket.Endpoint
	var src gopacket.Endpoint
	var dest gopacket.Endpoint
	for {
		p, err := packet.NextPacket()
		if err == io.EOF {
			break
		} else if err != nil {
			log.Println("Error:", err)
			continue
		}
		src, dest = p.NetworkLayer().NetworkFlow().Endpoints()
		if !((src == pastIPs[0]) && (dest == pastIPs[1])) && !((src == pastIPs[1]) && (dest == pastIPs[0])) {
			//means that this is a new flow
			fmt.Print(src, dest, "\n")
			pastIPs[0] = src
			pastIPs[1] = dest
		}
	}
}
