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

	tea "github.com/charmbracelet/bubbletea"
)

const (
	UDP = 0
	TCP = 1
)

type socketMap []socketKey
type socketMapMsg struct{ data *socketMap }

func getCStruct() *socketMap {

	var sockets C.int
	C.socketCount(&sockets)

	var socketInfo = make([]socketsDef, sockets)
	var goSocketInfo = make(socketMap, sockets)
	C.goSocketStructs(unsafe.Pointer(&socketInfo[0]), &sockets)

	for i := 0; i < len(socketInfo); i++ {
		socket := socketInfo[i]

		key := socketKey{
			ProcessName:  cStringToGo32(socket.ProcessName),
			DestIP:       cStringToGo16(socket.DestIPAddr),
			SrcPort:      strconv.Itoa(int(int32(socket.SourcePort))),
			DestPort:     strconv.Itoa(int(int32(socket.DestPort))),
			ConnType:     int32(socket.Connection_type),
			CreationTime: time.Unix(0, int64(socket.CreationTime)),
			Pid:          strconv.Itoa(int(int32(socket.Pid))),
		}
		goSocketInfo[i] = key

	}
	return &goSocketInfo
}

func cStringToGo32(b [32]int8) string {
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

func cStringToGo16(b [16]int8) string {
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
func getCStructCmd() tea.Cmd {
	return func() tea.Msg {
		return socketMapMsg{data: getCStruct()}
	}
}
