#include "customTypes.c"
#include "customTypes.h"

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

void filterSockets(enum filter_listening listening_filter,
                   enum filter_local local_filter,
                   struct socketInfo *socketData, int *socketNum);

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

  info->local = 0;
  info->pid = socket_info->pid;
  proc_name(socket_info->pid, info->processName, sizeof(processName_t));
  info->processName[31] = '\0';
  /* Initialize destIPAddr safely */
  strcpy(info->destIPAddr, "0");
  /* TCP / UDP discrimination */
  if (socket_info->socketinfo.psi.soi_kind == SOCKINFO_IN) {

    info->socket_type = UDP;
    info->sourcePort =
        ntohs(socket_info->socketinfo.psi.soi_proto.pri_in.insi_lport);
    info->destPort =
        ntohs(socket_info->socketinfo.psi.soi_proto.pri_in.insi_fport);

    /* Unconnected or listening */
    if (info->destPort == 0) {
      info->listening = 1;
      return;
    }

    if (socket_info->socketinfo.psi.soi_proto.pri_in.insi_vflag == INI_IPV4) {
      info->listening = 0;

      char ipStr[INET_ADDRSTRLEN];
      convAddrIpv4(socket_info->socketinfo.psi.soi_proto.pri_in.insi_faddr
                       .ina_46.i46a_addr4.s_addr,
                   ipStr);

      strlcpy(info->destIPAddr, ipStr,
              sizeof(info->destIPAddr)); // local detection
      struct in_addr v4;
      if (inet_pton(AF_INET, info->destIPAddr, &v4) == 1) {
        if ((ntohl(v4.s_addr) & 0xff000000) == 0x7f000000) {
          info->local = 1;
          return;
        }
      }

    } else if (socket_info->socketinfo.psi.soi_proto.pri_in.insi_vflag ==
               INI_IPV6) {

      info->listening = 0;

      char ipStr[INET6_ADDRSTRLEN];
      convAddrIpv6(socket_info->socketinfo.psi.soi_proto.pri_in.insi_faddr.ina_6
                       .__u6_addr.__u6_addr8,
                   ipStr);

      strlcpy(info->destIPAddr, ipStr,
              sizeof(info->destIPAddr)); // local detection
      struct in6_addr v6;
      if (inet_pton(AF_INET6, info->destIPAddr, &v6) == 1) {
        if (IN6_IS_ADDR_LOOPBACK(&v6)) {
          info->local = 1;
          return;
        }
      }
    }

  } else if (socket_info->socketinfo.psi.soi_kind == SOCKINFO_TCP) {

    info->socket_type = TCP;
    info->sourcePort = ntohs(
        socket_info->socketinfo.psi.soi_proto.pri_tcp.tcpsi_ini.insi_lport);
    info->destPort = ntohs(
        socket_info->socketinfo.psi.soi_proto.pri_tcp.tcpsi_ini.insi_fport);

    /* LISTEN sockets */
    if (info->destPort == 0) {
      info->listening = 1;
      return;
    }

    if (socket_info->socketinfo.psi.soi_proto.pri_tcp.tcpsi_ini.insi_vflag ==
        INI_IPV4) {

      info->listening = 0;
      info->connection_type = IPV4;

      char ipStr[INET_ADDRSTRLEN];
      convAddrIpv4(socket_info->socketinfo.psi.soi_proto.pri_tcp.tcpsi_ini
                       .insi_faddr.ina_46.i46a_addr4.s_addr,
                   ipStr);

      strlcpy(info->destIPAddr, ipStr,
              sizeof(info->destIPAddr)); // local detection
      // local detection
      struct in_addr v4;
      if (inet_pton(AF_INET, info->destIPAddr, &v4) == 1) {
        if ((ntohl(v4.s_addr) & 0xff000000) == 0x7f000000) {
          info->local = 1;
          return;
        }
      }

    } else if (socket_info->socketinfo.psi.soi_proto.pri_tcp.tcpsi_ini
                   .insi_vflag == INI_IPV6) {

      info->listening = 0;
      info->connection_type = IPV6;

      char ipStr[INET6_ADDRSTRLEN];
      convAddrIpv6(socket_info->socketinfo.psi.soi_proto.pri_tcp.tcpsi_ini
                       .insi_faddr.ina_6.__u6_addr.__u6_addr8,
                   ipStr);

      strlcpy(info->destIPAddr, ipStr,
              sizeof(info->destIPAddr)); // local detection
      struct in6_addr v6;
      if (inet_pton(AF_INET6, info->destIPAddr, &v6) == 1) {
        if (IN6_IS_ADDR_LOOPBACK(&v6)) {
          info->local = 1;
          return;
        }
      }
    }
  }
}

// utility functions for debugging

void logToFileInt(int x) {
  FILE *fp;
  fp = fopen("../debug.txt", "a");
  fprintf(fp, "\n socket num: %d", x);
  fclose(fp);
}

void logToFileStr(char *x) {
  FILE *fp;
  fp = fopen("./debug_c.txt", "a");
  fprintf(fp, "\n socket num: %s", x);
  fclose(fp);
}

void printSockets(struct socketInfo *socketData, int socketNum, int log) {
  remove("./debug_c.txt");
  for (int i = 0; i < socketNum; i++) {
    if (log) {
      char socketDesc[200];
      snprintf(socketDesc, 200,
               "process name: %s \n   - pid: %d \n   - connection type %d"
               "\n   - ip %s \n   - local port %d\n   - foreign port %d \n   - "
               "listening %d\n   - local %d\n",
               socketData[i].processName, socketData[i].pid,
               socketData[i].socket_type, socketData[i].destIPAddr,
               socketData[i].sourcePort, socketData[i].destPort,
               socketData[i].listening, socketData[i].local);
      logToFileStr(socketDesc);
    } else {
      printf("process name: %s \n   - pid: %d \n   - connection type %d"
             "\n   - ip %s \n   - local port %d\n   - foreign port %d \n   - "
             "listening %d\n   - local %d\n",
             socketData[i].processName, socketData[i].pid,
             socketData[i].socket_type, socketData[i].destIPAddr,
             socketData[i].sourcePort, socketData[i].destPort,
             socketData[i].listening, socketData[i].local);
      printf("about to log socket to file");
    }
  }
}

void goSocketStructs(void *goSocketData, int *socketNum) {
  int amtPids;

  pid_t *allPidsArr = getPids(&amtPids);

  pidsSocketsCountInternal(socketNum, &amtPids, allPidsArr);

  struct socketinfo_andpid *socketDataArr = (struct socketinfo_andpid *)malloc(
      *socketNum * sizeof(struct socketinfo_andpid));
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

  struct socketInfo *socketData = (struct socketInfo *)goSocketData;
  for (int i = 0; i < *socketNum; i++) {
    descSocket(&socketData[i], &socketDataArr[i]);
  }
  // printf("about to filter sockets\n");
  // printf("pre: %d\n", *socketNum);
  // filterSockets(NO_LISTEN, BOTH_LOCAL_STATES, socketData, socketNum);
  // printf("post: %d\n", *socketNum);
}

void removeElement(struct socketInfo *socketArr, int *size, int index) {
  socketArr[index] = socketArr[*size - 1];
  (*size)--;
}

// sorts sockets
void filterSockets(enum filter_listening listening_filter,
                   enum filter_local local_filter,
                   struct socketInfo *socketData, int *socketNum) {

  int write = 0;
  for (int read = 0; read < *socketNum; read++) {
    int keep = 1;
    if (local_filter != BOTH_LOCAL_STATES &&
        socketData[read].local != local_filter) {
      keep = 0;
    } else if (listening_filter != BOTH_LISTEN_STATES &&
               socketData[read].listening != listening_filter) {
      keep = 0;
    }
    if (keep) {
      if (write != read) {
        //    printf("listening = %d\n", socketData[read].listening);
        socketData[write] = socketData[read];
      }
      write++;
    }
  }
  // printf("%d", write);
  *socketNum = write;
  return;
}
