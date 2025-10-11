package main

import (
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

func test() {
	allPids, err := lc.ListAllPids(0)
	if err != nil {
		log.Fatal(err)
	}
	allFdInfo := make([][]proc_fdinfo, len(allPids))
	for i, pid := range allPids {

		pids := ProcPidInfo(pid)
		if pids != nil {
			allFdInfo[i] = pids
		}
	}
	//fmt.Print(allFdInfo[rand.Int31n(int32(len(allFdInfo)))])

}
