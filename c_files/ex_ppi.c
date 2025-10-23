#include "./customTypes.h"
#include <arpa/inet.h>
#include <libproc.h>
#include <netinet/in.h>
#include <stdbool.h>
#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <sys/proc_info.h>
#include <unistd.h>

struct proc_fdinfo *getFDs(pid_t pid, int *amtfds) {
  // get buffer size for file descriptors
  int bufferSize = proc_pidinfo(pid, PROC_PIDLISTFDS, 0, NULL, 0);
  if (bufferSize <= 0) {
    return NULL;
  }

  // reserve memory
  struct proc_fdinfo *fdInfo = (struct proc_fdinfo *)malloc(bufferSize);
  if (fdInfo == NULL) {
    perror("malloc");
    return NULL;
  }

  // read file descriptors
  int bytesRead = proc_pidinfo(pid, PROC_PIDLISTFDS, 0, fdInfo, bufferSize);
  if (bytesRead <= 0) {
    free(fdInfo);
    return NULL;
  }

  // optional: print sockets
  for (int i = 0; i < (bytesRead / PROC_PIDLISTFD_SIZE); i++) {
    if (fdInfo[i].proc_fdtype == PROX_FDTYPE_SOCKET) {
      // printf("PID %d has socket FD %d\n", pid, fdInfo[i].proc_fd);
    }
  }
  *amtfds = bytesRead / PROC_PIDLISTFD_SIZE;
  return fdInfo;
}

pid_t *getPids(int *amtPids) {
  int pidCount = proc_listpids(PROC_ALL_PIDS, 0, NULL, 0);
  if (pidCount <= 0) {
    perror("proc_listpids (count)");
    return NULL;
  }

  unsigned long pidsBufSize = sizeof(pid_t) * (unsigned long)pidCount;
  pid_t *pids = (pid_t *)malloc(pidsBufSize);
  if (!pids) {
    perror("malloc");
    return NULL;
  }

  int bytesUsed = proc_listpids(PROC_ALL_PIDS, 0, pids, (int)pidsBufSize);
  if (bytesUsed <= 0) {
    perror("proc_listpids (data)");
    free(pids);
    return NULL;
  }

  int numPids = bytesUsed / sizeof(pid_t);
  int newArrSize = 0;

  for (int i = 0; i < numPids; i++) {
    if (pids[i] != 0)
      newArrSize++;
  }

  pid_t *nonEmptyPids = (pid_t *)malloc(newArrSize * sizeof(pid_t));
  int arrTracker = 0;
  for (int i = 0; i < numPids; i++) {
    if (pids[i] != 0) {
      nonEmptyPids[arrTracker++] = pids[i];
    }
  }

  *amtPids = newArrSize;
  free(pids);
  return nonEmptyPids;
}

int getSocketData(pid_t pid, int fd, struct socket_fdinfo *socketInfo) {
  int bytesRead = proc_pidfdinfo(pid, fd, PROC_PIDFILEPORTSOCKETINFO,
                                 socketInfo, sizeof(struct socket_fdinfo));
  return bytesRead;
}

void convAddrIpv4(in_addr_t addr, char *ipStr) {
  struct in_addr address;
  address.s_addr = addr;
  inet_ntop(AF_INET, &address, ipStr, INET_ADDRSTRLEN);
}

void convAddrIpv6(const uint8_t addr[16], char *ipStr) {
  struct in6_addr address;
  memcpy(&address, addr, sizeof(address));
  inet_ntop(AF_INET6, &address, ipStr, INET6_ADDRSTRLEN);
}

struct socketinfo_andpid {
  pid_t pid;
  int fd;
  struct socket_fdinfo socketinfo;
};

int main(int argc, char *argv[]) {

  int amtPids = 0;

  pid_t *allPidsArr = getPids(&amtPids);

  // printf("Total PIDs: %d\n", amtPids);

  int socketNum = 0;
  for (int i = 0; i < amtPids; i++) {
    int amtFds = 0;
    struct proc_fdinfo *fdinfo = getFDs(allPidsArr[i], &amtFds);
    // just a quick check for sockets
    for (int j = 0; j < amtFds; j++) {
      if (fdinfo[j].proc_fdtype == PROX_FDTYPE_SOCKET) {
        //  printf("PID %d: socket detected\n", allPidsArr[i]);
        socketNum++;
        /*
        struct socket_fdinfo socketInfo;
        int bytesRead = getSocketData(allPidsArr[i], fdinfo[j].proc_fd,
        &socketInfo); printf("bytes read: %d \n", bytesRead);
        */
      }
    }
  }
  struct socketinfo_andpid *socketDataArr = (struct socketinfo_andpid *)malloc(
      socketNum * sizeof(struct socketinfo_andpid));
  int arrCounter = 0;
  for (int i = 0; i < amtPids; i++) {

    int amtFds = 0;
    struct proc_fdinfo *fdinfo = getFDs(allPidsArr[i], &amtFds);
    for (int j = 0; j < amtFds; j++) {
      if (fdinfo[j].proc_fdtype == PROX_FDTYPE_SOCKET) {
        struct socket_fdinfo socketInfo;
        socketDataArr[arrCounter].fd = fdinfo[j].proc_fd;
        socketDataArr[arrCounter].pid = allPidsArr[i];
        int bytesRead = getSocketData(allPidsArr[i], fdinfo[j].proc_fd,
                                      &socketDataArr[arrCounter].socketinfo);
        arrCounter++;
        //  printf("bytes read: %d \n", bytesRead);
      }
    }
  }
  struct socketInfo *goSocketData =
      (struct socketInfo *)malloc(socketNum * sizeof(struct socketInfo));

  for (int i = 0; i < socketNum; i++) {
    struct socketInfo info;
    info.pid = socketDataArr[i].pid;
    proc_name(socketDataArr[i].pid, info.processName, sizeof(processName_t));
    /*
    if (socketDataArr[i].socketinfo.psi.soi_kind == SOCKINFO_IN) {
         uint16_t lport =
    ntohs(socketDataArr[i].socketinfo.psi.soi_proto.pri_in.insi_lport); uint16_t
    fport = ntohs(socketDataArr[i].socketinfo.psi.soi_proto.pri_in.insi_fport);
        printf("UDP local port %d to foreign port %d on pid %d\n",
            ntohs(socketDataArr[i].socketinfo.psi.soi_proto.pri_in.insi_lport),
            ntohs(socketDataArr[i].socketinfo.psi.soi_proto.pri_in.insi_fport),
            socketDataArr[i].pid);

      if (socketDataArr[i].socketinfo.psi.soi_proto.pri_in.insi_vflag ==
    INI_IPV4) { char ipStr[INET_ADDRSTRLEN];
            convAddrIpv4(socketDataArr[i].socketinfo.psi.soi_proto.pri_in.insi_faddr.ina_46.i46a_addr4.s_addr,
    ipStr); printf("    Remote IPv4: %s:%u\n", ipStr, fport); } else if
    (socketDataArr[i].socketinfo.psi.soi_proto.pri_in.insi_vflag == INI_IPV6) {
            char ipStr[INET6_ADDRSTRLEN];
            convAddrIpv6(socketDataArr[i].socketinfo.psi.soi_proto.pri_in.insi_faddr.ina_6.__u6_addr.__u6_addr8,
    ipStr); printf("    Remote IPv6: [%s]:%u\n", ipStr, fport);
        }
    } else if (socketDataArr[i].socketinfo.psi.soi_kind == SOCKINFO_TCP) {
        uint16_t lport =
    ntohs(socketDataArr[i].socketinfo.psi.soi_proto.pri_tcp.tcpsi_ini.insi_lport);
        uint16_t fport =
    ntohs(socketDataArr[i].socketinfo.psi.soi_proto.pri_tcp.tcpsi_ini.insi_fport);

        printf("TCP local port %d to foreign port %d on pid %d\n", lport, fport,
    socketDataArr[i].pid);


        if
    (socketDataArr[i].socketinfo.psi.soi_proto.pri_tcp.tcpsi_ini.insi_vflag ==
    INI_IPV4) { char ipStr[INET_ADDRSTRLEN];
            convAddrIpv4(socketDataArr[i].socketinfo.psi.soi_proto.pri_tcp.tcpsi_ini.insi_faddr.ina_46.i46a_addr4.s_addr,
    ipStr); printf("Remote IPv4: %s:%u\n", ipStr, fport);
        }
        else if
    (socketDataArr[i].socketinfo.psi.soi_proto.pri_tcp.tcpsi_ini.insi_vflag ==
    INI_IPV6) { struct in6_addr faddr6 =
    socketDataArr[i].socketinfo.psi.soi_proto.pri_tcp.tcpsi_ini.insi_faddr.ina_6;
            char ipStr[INET6_ADDRSTRLEN];
            inet_ntop(AF_INET6, &faddr6, ipStr, sizeof(ipStr));

            printf("Remote IPv6: [%s]:%u\n", ipStr, fport);
        }
    }
        */
    if (socketDataArr[i].socketinfo.psi.soi_kind == SOCKINFO_IN) {

      info.socket_type = UDP;
      info.sourcePort =
          ntohs(socketDataArr[i].socketinfo.psi.soi_proto.pri_in.insi_lport);
      info.destPort =
          ntohs(socketDataArr[i].socketinfo.psi.soi_proto.pri_in.insi_fport);

      if (socketDataArr[i].socketinfo.psi.soi_proto.pri_in.insi_vflag ==
          INI_IPV4) {

        char ipStr[INET_ADDRSTRLEN];
        convAddrIpv4(socketDataArr[i]
                         .socketinfo.psi.soi_proto.pri_in.insi_faddr.ina_46
                         .i46a_addr4.s_addr,
                     ipStr);
        strncpy(info.destIPAddr, ipStr, sizeof(ipAddr));
      } else if (socketDataArr[i].socketinfo.psi.soi_proto.pri_in.insi_vflag ==
                 INI_IPV6) {
        char ipStr[INET6_ADDRSTRLEN];
        convAddrIpv6(socketDataArr[i]
                         .socketinfo.psi.soi_proto.pri_in.insi_faddr.ina_6
                         .__u6_addr.__u6_addr8,
                     ipStr);
        strncpy(info.destIPAddr, ipStr, sizeof(ipAddr));
      }
    } else if (socketDataArr[i].socketinfo.psi.soi_kind == SOCKINFO_TCP) {
      info.socket_type = TCP;
      info.sourcePort =
          ntohs(socketDataArr[i]
                    .socketinfo.psi.soi_proto.pri_tcp.tcpsi_ini.insi_lport);
      info.destPort =
          ntohs(socketDataArr[i]
                    .socketinfo.psi.soi_proto.pri_tcp.tcpsi_ini.insi_fport);

      if (socketDataArr[i]
              .socketinfo.psi.soi_proto.pri_tcp.tcpsi_ini.insi_vflag ==
          INI_IPV4) {
        info.connection_type = IPV4;
        char ipStr[INET_ADDRSTRLEN];
        convAddrIpv4(socketDataArr[i]
                         .socketinfo.psi.soi_proto.pri_tcp.tcpsi_ini.insi_faddr
                         .ina_46.i46a_addr4.s_addr,
                     ipStr);
        strncpy(info.destIPAddr, ipStr, sizeof(ipAddr));
      } else if (socketDataArr[i]
                     .socketinfo.psi.soi_proto.pri_tcp.tcpsi_ini.insi_vflag ==
                 INI_IPV6) {
        info.connection_type = IPV6;
        char ipStr[INET6_ADDRSTRLEN];
       convAddrIpv6((const uint8_t *)&socketDataArr[i]
                 .socketinfo.psi.soi_proto.pri_tcp.tcpsi_ini.insi_faddr.ina_6,
             ipStr);
        strncpy(info.destIPAddr, ipStr, sizeof(ipAddr));
      }
    }


info.local = 0;
info.listening = 0;

// Check first three characters against LOCAL_IPS
for (int i = 0; i < LOCAL_IPS_LEN; i++) {
    if (strncmp(info.destIPAddr, LOCAL_IPS[i], strlen(LOCAL_IPS[i])) == 0) {
        info.local = 1;
        break;  // found a match, no need to check further
    }
}

// Check first character against listening prefixes
if (info.destIPAddr[0] == LISTENING_IP_PREFIX[0] ||
    info.destIPAddr[0] == LISTENING_IP_PREFIX2[0]) {
    info.listening = 1;
}


    goSocketData[i] = info;
   
  }

  for (int i = 0; i < socketNum; i++) {
    if ((goSocketData[i].socket_type == TCP) && (goSocketData[i].local != 1) && (goSocketData[i].listening != 1)) {
  printf("process name: %s \n   - pid: %d \n   - connection type %d\n  - ip %s \n", goSocketData[i].processName, goSocketData[i].pid, goSocketData[i].socket_type, goSocketData[i].destIPAddr);
    }
  }

  return 0;



}
