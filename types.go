package main

import (
	"time"

	"github.com/google/gopacket"
)

type socketKey struct {
	ProcessName  string
	DestIP       string
	SrcPort      string
	DestPort     string
	ConnType     int32
	CreationTime time.Time
	Pid          string
	outgoing     int
	incoming     int
}

type PacketInfo struct {
	packetProtocol gopacket.LayerType
	sourceIP       string
	destIP         string
	sourcePort     string
	destPort       string
	eof            bool
	appData        []byte
	packetSize     int
	Timestamp      time.Time
}
