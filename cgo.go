package main

/*
#cgo CFLAGS: -I./c_files
#include "customTypes.h"
#include "ex_ppi.c"   // include the C implementation directly
*/
import "C"
import "fmt"

func getCStruct() {
	//first get amt of sockets

	var sockets C.int
	C.socketCount(&sockets)
	fmt.Print(sockets)
}
