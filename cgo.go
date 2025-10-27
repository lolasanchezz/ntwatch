package main

/*
#cgo CFLAGS: -I./c_files
#include "customTypes.h"
#include "customTypes.c"
#include "ex_ppi.c"   // include the C implementation directly
*/
import "C"
import (
	"fmt"
	"unsafe"
)

type socketKey struct {
	ProcessName string
	DestIP      string
	Port        int32
}

type socketMap map[socketKey]*socketsDef

func getCStruct() *socketMap {
	//first get amt of sockets

	var sockets C.int
	C.socketCount(&sockets)
	fmt.Print(sockets)

	var socketInfo = make([]socketsDef, sockets)
	var goSocketInfo = make(socketMap, sockets)
	C.goSocketStructs(unsafe.Pointer(&socketInfo[0]), sockets)
	for _, socket := range socketInfo {
		key := socketKey{
			ProcessName: C.GoString((*C.char)(unsafe.Pointer(&socket.ProcessName[0]))),
			DestIP:      C.GoString((*C.char)(unsafe.Pointer(&socket.DestIPAddr[0]))),
			Port:        int32(socket.SourcePort),
		}
		goSocketInfo[key] = &socket

	}
	return &goSocketInfo
}
