package main

const (
	IF_NAMESIZE     = 16
	SOCK_MAXADDRLEN = 128 // macOS typically 128
	MAX_KCTL_NAME   = 96
)

const (
	PROX_FDTYPE_ATALK = iota
	PROX_FDTYPE_VNODE
	PROX_FDTYPE_SOCKET
	PROX_FDTYPE_PSHM
	PROX_FDTYPE_PSEM
	PROX_FDTYPE_KQUEUE
	PROX_FDTYPE_PIPE
	PROX_FDTYPE_FSEVENTS
	PROX_FDTYPE_NETPOLICY
	PROX_FDTYPE_CHANNEL
	PROX_FDTYPE_NEXUS
)

// Equivalent to C's struct proc_fileinfo
type ProcFileInfo struct {
	FiOpenFlags  uint32
	FiStatus     uint32
	FiOffset     int64 // off_t
	FiType       int32
	FiGuardFlags uint32
}

// Equivalent to C's struct vinfo_stat
type VinfoStat struct {
	VstDev           uint32
	VstMode          uint16
	VstNlink         uint16
	VstIno           uint64
	VstUid           uint32
	VstGid           uint32
	VstAtime         int64
	VstAtimeNsec     int64
	VstMtime         int64
	VstMtimeNsec     int64
	VstCtime         int64
	VstCtimeNsec     int64
	VstBirthtime     int64
	VstBirthtimeNsec int64
	VstSize          int64
	VstBlocks        int64
	VstBlksize       int32
	VstFlags         uint32
	VstGen           uint32
	VstRdev          uint32
	VstQspare        [2]int64
}

// Equivalent to C's struct sockbuf_info
type SockbufInfo struct {
	SbiCc    uint32
	SbiHiwat uint32
	SbiMbcnt uint32
	SbiMbmax uint32
	SbiLowat uint32
	SbiFlags int16
	SbiTimeo int16
}

// struct in4in6_addr
type In4In6Addr struct {
	Pad32 [3]uint32
	Addr4 [4]byte // struct in_addr
}

// struct in_sockinfo
type InSockInfo struct {
	InsiFport  int32
	InsiLport  int32
	InsiGencnt uint64
	InsiFlags  uint32
	InsiFlow   uint32
	InsiVflag  byte
	InsiIPTtl  byte
	Rfu1       uint32

	// addresses
	InsiFaddr [16]byte
	InsiLaddr [16]byte

	// v4
	InsiV4Tos byte

	// v6
	InsiV6Hlim  byte
	InsiV6Cksum int32
	InsiV6Ifidx uint16
	InsiV6Hops  int16
}

// struct tcp_sockinfo
const TSI_T_NTIMERS = 4

type TcpSockInfo struct {
	TcpsiIni   InSockInfo
	TcpsiState int32
	TcpsiTimer [TSI_T_NTIMERS]int32
	TcpsiMss   int32
	TcpsiFlags uint32
	Rfu1       uint32
	TcpsiTp    uint64 // pointer
}

// struct un_sockinfo
type UnSockInfo struct {
	UnsiConnSo  uint64
	UnsiConnPcb uint64
	UnsiAddr    [SOCK_MAXADDRLEN]byte
	UnsiCaddr   [SOCK_MAXADDRLEN]byte
}

// struct ndrv_info
type NdrvInfo struct {
	NdrvsiIfFamily uint32
	NdrvsiIfUnit   uint32
	NdrvsiIfName   [IF_NAMESIZE]byte
}

// struct kern_event_info
type KernEventInfo struct {
	KesiVendorCodeFilter uint32
	KesiClassFilter      uint32
	KesiSubclassFilter   uint32
}

// struct kern_ctl_info
type KernCtlInfo struct {
	KcsiID          uint32
	KcsiRegUnit     uint32
	KcsiFlags       uint32
	KcsiRecvBufSize uint32
	KcsiSendBufSize uint32
	KcsiUnit        uint32
	KcsiName        [MAX_KCTL_NAME]byte
}

// struct vsock_sockinfo
type VsockSockInfo struct {
	LocalCID   uint32
	LocalPort  uint32
	RemoteCID  uint32
	RemotePort uint32
}

// struct socket_info
type SocketInfo struct {
	SoiStat     VinfoStat
	SoiSo       uint64
	SoiPcb      uint64
	SoiType     int32
	SoiProtocol int32
	SoiFamily   int32
	SoiOptions  int16
	SoiLinger   int16
	SoiState    int16
	SoiQlen     int16
	SoiIncqlen  int16
	SoiQlimit   int16
	SoiTimeo    int16
	SoiError    uint16
	SoiOobmark  uint32
	SoiRcv      SockbufInfo
	SoiSnd      SockbufInfo
	SoiKind     int32
	Rfu1        uint32

	// Union of protocol-specific info
	SoiProto [512]byte // large enough buffer for any of the unions
}

// struct socket_fdinfo
type SocketFDInfo struct {
	Pfi ProcFileInfo
	Psi SocketInfo
}
