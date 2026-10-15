package main

/*
#cgo CFLAGS: -I./c_files
#include "bridge.h"
*/
import "C"
import (
	"strconv"
	"time"
	"unsafe"
)

const (
	UDP = 0
	TCP = 1
)

type socketMap []socketKey
type socketMapMsg struct{ data socketMap }

func getCStruct() socketMap {

	var sockets C.int
	C.socketCount(&sockets)
	if sockets == 0 {
		return nil
	}

	// sockets can open between counting and filling, so leave some room. the C side
	// never writes past capacity and tells us how many it actually filled
	capacity := sockets + sockets/4 + 16
	var socketInfo = make([]socketsDef, capacity)
	C.goSocketStructs(unsafe.Pointer(&socketInfo[0]), capacity, &sockets)
	socketInfo = socketInfo[:sockets]
	var goSocketInfo = make(socketMap, sockets)

	for i := 0; i < len(socketInfo); i++ {
		socket := socketInfo[i]

		key := socketKey{
			ProcessName:  cString(socket.ProcessName[:]),
			DestIP:       cString(socket.DestIPAddr[:]),
			Proto:        int32(socket.Socket_type),
			SrcPort:      strconv.Itoa(int(int32(socket.SourcePort))),
			DestPort:     strconv.Itoa(int(int32(socket.DestPort))),
			ConnType:     int32(socket.Connection_type),
			CreationTime: time.Unix(0, int64(socket.CreationTime)),
			Pid:          strconv.Itoa(int(int32(socket.Pid))),
		}
		goSocketInfo[i] = key

	}
	return goSocketInfo
}

// reads a null-terminated C char array
func cString(b []int8) string {
	n := 0
	for ; n < len(b); n++ {
		if b[n] == 0 { //if the string has ended
			break
		}
	}
	bytes := make([]byte, n)
	for i := 0; i < n; i++ {
		bytes[i] = byte(b[i])
	}
	return string(bytes)
}

func getCStructCmd() Cmd {
	return func() Msg {
		return socketMapMsg{data: getCStruct()}
	}
}
