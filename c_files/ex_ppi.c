
#include <stdbool.h>
#include <stdio.h>
#include <stdlib.h>
#include <unistd.h>
#include <libproc.h>
#include <sys/proc_info.h>
#include <arpa/inet.h>


struct proc_fdinfo* getFDs(pid_t pid, int *amtfds) {
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
    *amtfds = bytesRead/PROC_PIDLISTFD_SIZE;
    return fdInfo; 
}


pid_t* getPids(int *amtPids) {
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
        if (pids[i] != 0) newArrSize++;
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
    int bytesRead = proc_pidfdinfo(pid, fd, PROC_PIDFILEPORTSOCKETINFO, socketInfo, sizeof(struct socket_fdinfo)); 
    return bytesRead;
}





int main(int argc, char *argv[]) { 
    
    int amtPids = 0;

    pid_t* allPidsArr = getPids(&amtPids);
   
    //printf("Total PIDs: %d\n", amtPids);

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
            int bytesRead = getSocketData(allPidsArr[i], fdinfo[j].proc_fd, &socketInfo);         
            printf("bytes read: %d \n", bytesRead);
            */
            }   
        }
    }
     struct socket_fdinfo* socketDataArr = (struct socket_fdinfo *)malloc(socketNum * sizeof(struct socket_fdinfo));
     int arrCounter = 0;
     for (int i = 0; i < amtPids; i++) {
       
        int amtFds = 0;
        struct proc_fdinfo *fdinfo = getFDs(allPidsArr[i], &amtFds);
        for (int j = 0; j < amtFds; j++) {
            if (fdinfo[j].proc_fdtype == PROX_FDTYPE_SOCKET) {
                struct socket_fdinfo socketInfo;
                int bytesRead = getSocketData(allPidsArr[i], fdinfo[j].proc_fd, &socketDataArr[arrCounter]);
                arrCounter++;        
              //  printf("bytes read: %d \n", bytesRead);
            }
        }

    
    
}

//int socketType = socketDataArr[150].psi.soi_kind;
/*
switch (socketType) { 
    case SOCKINFO_GENERIC:
        printf("generic socket!");
        break;
    case SOCKINFO_IN:
        printf("networking port!");
        break;
    case SOCKINFO_TCP: 
        printf("this is tcp!");
        break;
    case SOCKINFO_UN:
        printf("this is a unix socket!");
        break;
    case SOCKINFO_NDRV:
        printf("this is a pf_ndrv socket!");
        break;
    case SOCKINFO_KERN_EVENT:
        printf("this is a kernel event socket!");
        break;
    case SOCKINFO_KERN_CTL:
        printf("this is a kernel control socket!");
        break;
    case SOCKINFO_VSOCK:
        printf("this is a virtual socket!");
        break;
}
*/
for (int i = 0; i < socketNum; i++) {
    if (socketDataArr[i].psi.soi_kind == SOCKINFO_IN) {
       printf("port %d \n", ntohs(socketDataArr[i].psi.soi_proto.pri_in.insi_fport));
    } else if (socketDataArr[i].psi.soi_kind == SOCKINFO_TCP) {
        printf("tcp port %d \n", ntohs(socketDataArr[i].psi.soi_proto.pri_tcp.tcpsi_ini.insi_fport));
    }
}






return 0;
}







