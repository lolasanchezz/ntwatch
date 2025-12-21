package main

/*
import (
	"fmt"
	"log"
	"math/rand"
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

func randomIndArr[T any](arr []T) int {
	return int(rand.Int31n(int32(len(arr))))
}
func randomIndMap[M ~map[K]V, K comparable, V any](Map M) any {
	ind := int(rand.Int31n(int32(len(Map))))
	i := 0
	for key, _ := range Map {
		if i == ind {
			fmt.Print(key)
			return key

		}
		i++
	}
	return nil
}

func test() {
	allPids, err := lc.ListAllPids(0)
	if err != nil {
		log.Fatal(err)
	}

	socketFds := make(map[lc.Pid][]proc_fdinfo, len(allPids))
	allFdInfo := make(map[lc.Pid][]proc_fdinfo, len(allPids))

	for _, pid := range allPids {
		fds := ProcPidInfo(pid)
		if fds == nil {
			continue
		}
		allFdInfo[pid] = fds
		for _, fd := range fds {
			if fd.proc_fdtype == PROX_FDTYPE_SOCKET {
				socketFds[pid] = append(socketFds[pid], fd)
			}
		}
	}

	socketFdInfo := make(map[lc.Pid][]SocketFDInfo, len(socketFds))
	for pid, sockArr := range socketFds {
		socketFdInfo[pid] = make([]SocketFDInfo, len(sockArr))
	}

	for pid, sockArr := range socketFds {
		for i, fd := range sockArr {
			var info SocketFDInfo
			_ = int(unsafe.Sizeof(info))
			//first get the size of the buffer

			size, err := lc.RawProcPidFDInfo(pid, int(fd.proc_fd), lc.ProcPidfdsocketinfo,
				unsafe.Pointer(nil), 0)
			fmt.Printf(" \n err: %v", err)
			if err != nil {

				if err.Error() == "cannot allocate memory" {
					continue // this happens all the time - enomem - happens sporadically
				}
				log.Printf("pid %d fd %d: %v", pid, fd.proc_fd, err)
				continue
			}

			bytesRead, err := lc.RawProcPidFDInfo(pid, int(fd.proc_fd), lc.ProcPidfdsocketinfo,
				unsafe.Pointer(&socketFdInfo[pid][i]), size)

			// Store the successfully read info
			fmt.Printf(" \n bytes read: %d", bytesRead)
		}
	}

}
*/
