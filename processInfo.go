package main

import (
	"fmt"
	"log"
	"unsafe"

	lc "github.com/tejasmanohar/go-libproc"
)

type proc_fdinfo struct {
	proc_fd     int32
	proc_fdtype uint32
}

func ProcPidInfo(pid lc.Pid) []proc_fdinfo {
	bufSize, err := lc.RawProcPidInfo(pid, lc.ProcPidlistfds, 0, unsafe.Pointer(nil), 0)
	if err != nil {
		log.Fatal(err)
	}

	if bufSize == 0 {
		return nil
	}

	// allocate based on buf size which is much more than real number of file descriptors
	maxFds := int(bufSize) / int(unsafe.Sizeof(proc_fdinfo{}))
	arr := make([]proc_fdinfo, maxFds)

	// get actual number of bytes written in
	actualBytes, err := lc.RawProcPidInfo(pid, lc.ProcPidlistfds, 0, unsafe.Pointer(&(arr[0])), bufSize)
	if err != nil {
		log.Fatal(err)
	}

	//trim slice down to real file descriptors and get rid of trailing zero structs
	return arr[:int(actualBytes)/int(unsafe.Sizeof(proc_fdinfo{}))]
}

//socket types
// TODO move to a different file

func test() {
	allPids, err := lc.ListAllPids(0)
	if err != nil {
		log.Fatal(err)
	}
	allFdInfo := make(map[lc.Pid][]proc_fdinfo, len(allPids))
	for _, pid := range allPids {

		pids := ProcPidInfo(pid)
		if pids != nil {
			allFdInfo[pid] = pids
		}
	}
	//fmt.Print(allFdInfo[rand.Int31n(int32(len(allFdInfo)))])
	sockets := make(map[lc.Pid][]SocketInfo, len(allPids))

	//populate sockets arr with pids
	for _, pid := range allPids {
		sockets[pid] = make([]SocketInfo, 5)
	}

	for pid, fdArr := range allFdInfo {
		for i, fd := range fdArr {
			if fd.proc_fdtype == PROX_FDTYPE_SOCKET {
				bytesWritten, err := lc.RawProcPidFDInfo(pid, int(fd.proc_fd), lc.ProcPidfdsocketinfo, unsafe.Pointer(&(sockets[pid][i])), int(unsafe.Sizeof(SocketInfo{})))
				if bytesWritten == 0 || err != nil {
					fmt.Printf("bytes written was 0, %n", err)
				}
			}

		}
	}

}
