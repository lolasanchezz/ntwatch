#include <stdio.h>
#include <stdlib.h>
#include <unistd.h>
#include <libproc.h>
#include <sys/proc_info.h>

struct proc_fdinfo* getFDs(pid_t pid) {
    
    //get buffer size for file descriptors
    printf("%d", pid);
    int bufferSize = proc_pidinfo(pid, PROC_PIDLISTFDS, 0, NULL, 0);
    if (bufferSize <= 0) {
        perror("proc_pidinfo (size)");
        exit(1);
        return NULL;
    }
    
    //reserve memory
    struct proc_fdinfo *fdInfo = (struct proc_fdinfo *)malloc(bufferSize);
    if (fdInfo == NULL) {
        perror("malloc");
        return NULL;
    }

    //reading in the bytes into fdInfo
    int bytesRead = proc_pidinfo(pid, PROC_PIDLISTFDS, 0, fdInfo, bufferSize);
    if (bytesRead <= 0) {
        perror("proc_pidinfo (data)");
        free(fdInfo);
        return NULL;
    }

    
    for (int i = 0; i < (bytesRead / PROC_PIDLISTFD_SIZE); i++) {
        if (fdInfo[i].proc_fdtype == 6) {
            printf("%d \n", pid);
        }
    }
    //use proc_pidfdinfo once file descriptors are obtained
    return NULL; 
}


pid_t* getPids() {

    int pidCount = proc_listpids(PROC_ALL_PIDS, 0, NULL, 0);
    unsigned long pidsBufSize = sizeof(pid_t) * (unsigned long)pidCount;
    pid_t *pids = (pid_t *)malloc(pidsBufSize);

    int bytesUsed = proc_listpids(PROC_ALL_PIDS, 0, pids, (int)pidsBufSize);
    int numPids = bytesUsed / sizeof(pid_t);

    for (int i = 0; i < numPids; i++) {
        if (pids[i] != 0) {  // filter out "empty" slots, why does proc_listpids do this?
            printf("PID: %d\n", pids[i]);
        }
    }

    return pids;
}





int main(int argc, char *argv[]) { 
    pid_t myPid;
    myPid = 94098;
  

    getPids();

}