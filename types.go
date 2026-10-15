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
	Proto        int32 // UDP or TCP
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
	outgoing       bool       // sent from this machine (vs received)
	broadcast      bool       // sent to a broadcast/multicast address, so no one process owns it
	hostnames      []hostname // names this packet reveals (dns answers, tls server names)
	readable       string     // readable text in the payload, if it isn't encrypted
}

// what the views get - exported + tagged so wails can hand it to the frontend as json

type PacketView struct {
	Protocol    string `json:"protocol"`
	SourceIP    string `json:"sourceIP"`
	DestIP      string `json:"destIP"`
	SourcePort  string `json:"sourcePort"`
	DestPort    string `json:"destPort"`
	Size        int    `json:"size"`
	Time        int64  `json:"time"`        // unix millis
	Outgoing    bool   `json:"outgoing"`    // sent from this machine (vs received)
	RemoteHost  string `json:"remoteHost"`  // the other end's hostname, if one was seen
	ProcessName string `json:"processName"` // empty if the packet didn't match a socket
	System      bool   `json:"system"`      // broadcast/multicast/ARP etc - traffic no process owns
	Pid         string `json:"pid"`
}

type SocketView struct {
	ProcessName string       `json:"processName"`
	Pid         string       `json:"pid"`
	Packets     []PacketView `json:"packets"`     // newest first
	Up          int64        `json:"up"`          // bytes sent in the last minute
	Down        int64        `json:"down"`        // bytes received in the last minute
	Connections int          `json:"connections"` // open connections right now (needs network statistics)
	MedianMs    float64      `json:"medianMs"`    // median tcp round trip, 0 if unknown
	ConnectedTo []string     `json:"connectedTo"` // hosts it has open connections to, named ones first
}

type Snapshot struct {
	Interface     string       `json:"interface"` // what's being captured on
	Capturing     bool         `json:"capturing"`
	Err           string       `json:"err"`
	RecentSockets []SocketView `json:"recentSockets"` // most recently active first
	Selected      *SocketView  `json:"selected"`      // the process picked for the detailed view, nil if there are none
	NewPackets    []PacketView `json:"newPackets"`    // since the last snapshot, oldest first
	Analytics     Analytics    `json:"analytics"`
}

type Analytics struct {
	FlowsAvailable  bool             `json:"flowsAvailable"` // latency + open ports need network statistics
	Bandwidth       []BandwidthPoint `json:"bandwidth"`      // one per second for the last minute, oldest first
	TopProcesses    []ProcessUsage   `json:"topProcesses"`   // most bytes in the last minute
	TopDestinations []Destination    `json:"topDestinations"`
	Plaintext       []PlaintextAlert `json:"plaintext"` // newest first
	Latency         []ProcessLatency `json:"latency"`   // slowest first
	OpenPorts       []OpenPort       `json:"openPorts"` // reachable from the network first
}

type BandwidthPoint struct {
	Time int64 `json:"time"` // unix millis, start of the second
	Up   int64 `json:"up"`   // bytes sent from this machine
	Down int64 `json:"down"` // bytes received
}

type ProcessUsage struct {
	ProcessName string `json:"processName"`
	Pid         string `json:"pid"`
	Up          int64  `json:"up"`
	Down        int64  `json:"down"`
}

type Destination struct {
	IP        string   `json:"ip"`
	Host      string   `json:"host"` // empty if no name was seen for it
	Bytes     int64    `json:"bytes"`
	Packets   int64    `json:"packets"`
	Processes []string `json:"processes"` // a few of the processes talking to it
}

type PlaintextAlert struct {
	ProcessName string `json:"processName"`
	Pid         string `json:"pid"`
	Remote      string `json:"remote"` // hostname, or ip if none was seen
	RemotePort  string `json:"remotePort"`
	Protocol    string `json:"protocol"`
	Outgoing    bool   `json:"outgoing"`
	Snippet     string `json:"snippet"` // the readable text, trimmed
	Count       int    `json:"count"`   // packets like this so far
	Time        int64  `json:"time"`    // unix millis of the latest
}

type ProcessLatency struct {
	ProcessName string  `json:"processName"`
	Pid         string  `json:"pid"`
	MedianMs    float64 `json:"medianMs"` // across its open tcp connections
	Connections int     `json:"connections"`
	Slowest     string  `json:"slowest"` // hostname or ip of its slowest connection
	SlowestMs   float64 `json:"slowestMs"`
}

type OpenPort struct {
	ProcessName string `json:"processName"`
	Pid         string `json:"pid"`
	Protocol    string `json:"protocol"`
	Port        string `json:"port"`
	Address     string `json:"address"` // what it's bound to
	Network     bool   `json:"network"` // reachable from other machines, not just this one
}
