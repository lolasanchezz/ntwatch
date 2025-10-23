#include "./customTypes.h"
#include <arpa/inet.h>
#include <libproc.h>
#include <netinet/in.h>
#include <stdbool.h>
#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <sys/_types/_pid_t.h>
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

void pidsSocketsCountInternal(int *numSockets, int *pidAmt,
                              pid_t *allPidsArrPtr) {

  int socketNum = 0;
  for (int i = 0; i < *pidAmt; i++) {
    int amtFds = 0;
    struct proc_fdinfo *fdinfo = getFDs(allPidsArrPtr[i], &amtFds);
    // checking for sockets!
    for (int j = 0; j < amtFds; j++) {
      if (fdinfo[j].proc_fdtype == PROX_FDTYPE_SOCKET) {
        socketNum++;
      }
    }
  }
  *numSockets = socketNum;
}

void socketCount(int *numSocketsPtr) {
  int amtPids;
  pid_t *allPidsArr = getPids(&amtPids);

  int socketNum = 0;
  for (int i = 0; i < amtPids; i++) {
    int amtFds = 0;
    struct proc_fdinfo *fdinfo = getFDs(allPidsArr[i], &amtFds);
    // checking for sockets!
    for (int j = 0; j < amtFds; j++) {
      if (fdinfo[j].proc_fdtype == PROX_FDTYPE_SOCKET) {
        socketNum++;
      }
    }
  }
  *numSocketsPtr = socketNum;
}

void descSocket(struct socketInfo *info,
                struct socketinfo_andpid *socket_info) {

  info->pid = socket_info->pid;
  proc_name(socket_info->pid, info->processName, sizeof(processName_t));

  if (socket_info->socketinfo.psi.soi_kind == SOCKINFO_IN) {

    info->socket_type = UDP;
    info->sourcePort =
        ntohs(socket_info->socketinfo.psi.soi_proto.pri_in.insi_lport);
    info->destPort =
        ntohs(socket_info->socketinfo.psi.soi_proto.pri_in.insi_fport);

    if (socket_info->socketinfo.psi.soi_proto.pri_in.insi_vflag == INI_IPV4) {

      char ipStr[INET_ADDRSTRLEN];
      convAddrIpv4(socket_info->socketinfo.psi.soi_proto.pri_in.insi_faddr
                       .ina_46.i46a_addr4.s_addr,
                   ipStr);
      strncpy(info->destIPAddr, ipStr, sizeof(ipAddr));
    } else if (socket_info->socketinfo.psi.soi_proto.pri_in.insi_vflag ==
               INI_IPV6) {
      char ipStr[INET6_ADDRSTRLEN];
      convAddrIpv6(socket_info->socketinfo.psi.soi_proto.pri_in.insi_faddr.ina_6
                       .__u6_addr.__u6_addr8,
                   ipStr);
      strncpy(info->destIPAddr, ipStr, sizeof(ipAddr));
    }
  } else if (socket_info->socketinfo.psi.soi_kind == SOCKINFO_TCP) {
    info->socket_type = TCP;
    info->sourcePort = ntohs(
        socket_info->socketinfo.psi.soi_proto.pri_tcp.tcpsi_ini.insi_lport);
    info->destPort = ntohs(
        socket_info->socketinfo.psi.soi_proto.pri_tcp.tcpsi_ini.insi_fport);

    if (socket_info->socketinfo.psi.soi_proto.pri_tcp.tcpsi_ini.insi_vflag ==
        INI_IPV4) {
      info->connection_type = IPV4;
      char ipStr[INET_ADDRSTRLEN];
      convAddrIpv4(socket_info->socketinfo.psi.soi_proto.pri_tcp.tcpsi_ini
                       .insi_faddr.ina_46.i46a_addr4.s_addr,
                   ipStr);
      strncpy(info->destIPAddr, ipStr, sizeof(ipAddr));
    } else if (socket_info->socketinfo.psi.soi_proto.pri_tcp.tcpsi_ini
                   .insi_vflag == INI_IPV6) {
      info->connection_type = IPV6;
      char ipStr[INET6_ADDRSTRLEN];
      convAddrIpv6((const uint8_t *)&socket_info->socketinfo.psi.soi_proto
                       .pri_tcp.tcpsi_ini.insi_faddr.ina_6,
                   ipStr);
      strncpy(info->destIPAddr, ipStr, sizeof(ipAddr));
    }
  }

  info->local = 0;
  info->listening = 0;

  // Check first three characters against LOCAL_IPS
  for (int i = 0; i < LOCAL_IPS_LEN; i++) {
    if (strncmp(info->destIPAddr, LOCAL_IPS[i], strlen(LOCAL_IPS[i])) == 0) {
      info->local = 1;
      break; // found a match, no need to check further
    }
  }

  // Check first character against listening prefixes
  if (info->destIPAddr[0] == LISTENING_IP_PREFIX[0] ||
      info->destIPAddr[0] == LISTENING_IP_PREFIX2[0]) {
    info->listening = 1;
  }
}

void printSockets(struct socketInfo *socketData, int socketNum) {

  for (int i = 0; i < socketNum; i++) {

    if ((socketData[i].socket_type == TCP) && (socketData[i].local != 1)) { //&&
      //   (goSocketData[i].listening != 1)) {
      printf("process name: %s \n   - pid: %d \n   - connection type %d"
             "\n   - ip %s \n   - local port %d\n   - foreign port %d \n\n",
             socketData[i].processName, socketData[i].pid,
             socketData[i].socket_type, socketData[i].destIPAddr,
             socketData[i].sourcePort, socketData[i].destPort);
    }
  }
}

void goSocketStructs(struct socketInfo *goSocketData, int socketNum) {
  int amtPids;

  pid_t *allPidsArr = getPids(&amtPids);

  pidsSocketsCountInternal(&socketNum, &amtPids, allPidsArr);

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
      }
    }
  }
  // previous declaration of go socket data for reference
  /*
  struct socketInfo *goSocketData =
      (struct socketInfo *)malloc(socketNum * sizeof(struct socketInfo));
*/
  for (int i = 0; i < socketNum; i++) {
    descSocket(&goSocketData[i], &socketDataArr[i]);
  }
}

int main(int argc, char *argv[]) {
  int amtPids;
  int socketNum;

  pid_t *allPidsArr = getPids(&amtPids);

  pidsSocketsCountInternal(&socketNum, &amtPids, allPidsArr);
  struct socketInfo *goSocketData =
      (struct socketInfo *)malloc(socketNum * sizeof(struct socketInfo));

  goSocketStructs(goSocketData, socketNum);

  // for testing purposes only
  printSockets(goSocketData, socketNum);

  return 0;
}
