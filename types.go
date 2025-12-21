package main

import "github.com/google/gopacket"

type socketKey struct {
	ProcessName string
	DestIP      string
	SrcPort     string
	DestPort    string
	ConnType    int32
}

type PacketInfo struct {
	packetProtocol gopacket.LayerType
	sourceIP       string
	destIP         string
	sourcePort     string
	destPort       string
	eof            bool
	appData        []byte
}
