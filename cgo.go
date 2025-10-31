package main

/*
#cgo CFLAGS: -I./c_files
#include "customTypes.h"
#include "customTypes.c"
#include "ex_ppi.c"   // include the C implementation directly
*/
import "C"
import (
	"unsafe"
)

type socketKey struct {
	ProcessName string
	DestIP      string
	SrcPort     int32
	DestPort    int32
	ConnType    int32
}

const (
	UDP = 0
	TCP = 1
)

type socketMap map[socketKey]*socketsDef

func getCStruct() *socketMap {
	//first get amt of sockets

	var sockets C.int
	C.socketCount(&sockets)

	var socketInfo = make([]socketsDef, sockets)
	var goSocketInfo = make(socketMap, sockets)
	C.goSocketStructs(unsafe.Pointer(&socketInfo[0]), sockets)

	for _, socket := range socketInfo {
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
