package main

/*
#cgo CFLAGS: -I./c_files
#include "bridge.h"
*/
import "C"
import (
	"unsafe"
)

const (
	UDP = 0
	TCP = 1
)

type socketMap map[socketKey]*socketsDef

func getCStruct() *socketMap {

	var sockets C.int
	C.socketCount(&sockets)

	var socketInfo = make([]socketsDef, sockets)
	var goSocketInfo = make(socketMap, sockets)
	C.goSocketStructs(unsafe.Pointer(&socketInfo[0]), &sockets)
	filteredSockets := int(sockets)
	for i := 0; i < filteredSockets; i++ {
		socket := socketInfo[i]
		key := socketKey{
			ProcessName: C.GoString((*C.char)(unsafe.Pointer(&socket.ProcessName[0]))),
			DestIP:      C.GoString((*C.char)(unsafe.Pointer(&socket.DestIPAddr[0]))),
			SrcPort:     int32(socket.SourcePort),
			DestPort:    int32(socket.DestPort),
			ConnType:    int32(socket.Connection_type),
		}
		goSocketInfo[key] = &socket

	}
	return &goSocketInfo
}
