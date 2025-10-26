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

func getCStruct() {
	//first get amt of sockets

	var sockets C.int
	C.socketCount(&sockets)
	fmt.Print(sockets)

	var socketInfo = make([]socketsDef, sockets)
	C.goSocketStructs(unsafe.Pointer(&socketInfo[0]), sockets)
	for _, socket := range socketInfo {
		name := C.GoString((*C.char)(unsafe.Pointer(&socket.ProcessName[0])))
		fmt.Printf("Process Name: %s\n", name)
		fmt.Printf("Listening: %d\n", socket.Listening)
		DestipAddr := C.GoString((*C.char)(unsafe.Pointer(&socket.DestIPAddr[0])))
		fmt.Printf("Dest Ip, %s\n", DestipAddr)
	}
}

func bytesToStr(arr [32]int8) string {
	b := make([]byte, 0, len(arr))
	for _, c := range arr {
		if c == 0 {
			break
		}
		b = append(b, byte(c))
	}

	return string(b)
}
