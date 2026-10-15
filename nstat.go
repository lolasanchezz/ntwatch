package main

/*
#cgo CFLAGS: -I./c_files -fblocks
#cgo LDFLAGS: -framework CoreFoundation
#include "nstat.c"
*/
import "C"
import (
	"fmt"
	"strconv"
)

// flows from NetworkStatistics - see c_files/nstat.c

type nstatStartedMsg struct{}
type nstatFailedMsg struct{ err error }

// a flow showed up or changed (sock is set), or went away (removed)
type flowMsg struct {
	id        uint64
	removed   bool
	sock      socketKey // shaped like a scanned socket, so everything downstream treats them the same
	localIP   string
	listening bool // tcp, waiting for connections
}

// a fresh round trip time for a tcp flow
type flowStatsMsg struct {
	id    uint64
	rttMs float64
}

func nstatStart() Msg {
	var errBuf [256]C.char
	if C.nstat_start(&errBuf[0], C.int(len(errBuf))) != 0 {
		return nstatFailedMsg{fmt.Errorf("network statistics: %s", C.GoString(&errBuf[0]))}
	}
	return nstatStartedMsg{}
}

// blocks until the next flow event, like sendPacketCmd does for packets
func nstatNextCmd() Cmd {
	return func() Msg {
		var ev C.struct_nstat_event
		C.nstat_next(&ev)

		if ev.kind == C.NSTAT_STATS {
			return flowStatsMsg{id: uint64(ev.id), rttMs: float64(ev.rttMs)}
		}
		msg := flowMsg{id: uint64(ev.id), removed: ev.kind == C.NSTAT_REMOVED}
		if msg.removed {
			return msg
		}
		msg.localIP = C.GoString(&ev.localAddr[0])
		msg.listening = ev.listening != 0
		msg.sock = socketKey{
			ProcessName: C.GoString(&ev.processName[0]),
			Pid:         strconv.Itoa(int(ev.pid)),
			SrcPort:     strconv.Itoa(int(ev.localPort)),
			DestIP:      C.GoString(&ev.remoteAddr[0]),
			DestPort:    strconv.Itoa(int(ev.remotePort)),
			Proto:       int32(ev.proto),
		}
		if ev.remotePort == 0 {
			msg.sock.DestIP = "0" // unconnected, written the way the socket scan writes it
		}
		return msg
	}
}
