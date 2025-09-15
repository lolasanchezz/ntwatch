#include <stdio.h>
#include <stdlib.h>
#include <unistd.h>
#include <libproc.h>
#include <sys/proc_info.h>
/*
int main(int argc, char *argv[]) {
    /*
    pid_t pid = getpid(); // Get current process ID
    
    // Step 1: Determine buffer size
    int bufferSize = proc_pidinfo(pid, PROC_PIDLISTFDS, 0, NULL, 0);
    if (bufferSize <= 0) {
        perror("proc_pidinfo (size)");
        return 1;
    }

    // Step 2: Allocate buffer and retrieve information
    struct proc_fdinfo *fdInfo = (struct proc_fdinfo *)malloc(bufferSize);
    if (fdInfo == NULL) {
        perror("malloc");
        return 1;
    }

    int bytesRead = proc_pidinfo(pid, PROC_PIDLISTFDS, 0, fdInfo, bufferSize);
    if (bytesRead <= 0) {
        perror("proc_pidinfo (data)");
        free(fdInfo);
        return 1;
    }

    int numberOfFDs = bytesRead / PROC_PIDLISTFD_SIZE;
    printf("Open file descriptors for PID %d:\n", pid);
    for (int i = 0; i < numberOfFDs; i++) {
        printf("  FD: %d, Type: %d\n", fdInfo[i].proc_fd, fdInfo[i].proc_fdtype);
    }

    free(fdInfo);
    return 0;
    
   getFDs(3010);
}
*/
struct proc_fdinfo* getFDs(pid_t pid) {
    /*
    //get buffer size for file descriptors
    printf("%d", pid);
    int bufferSize = proc_pidinfo(pid, PROC_PIDTASKALLINFO, 0, NULL, 0);
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
*/

    struct proc_fdinfo fds;

    int bytes = proc_pidinfo(pid, PROC_PIDLISTFDS, 0, &fds, sizeof(fds));
    if (bytes <= 0) {
        perror("proc_pidinfo");
    exit(1);
    }
    
    printf("%d,", fds);
    //use proc_pidfdinfo once file descriptors are obtained

}





int main(int argc, char *argv[]) {

    getFDs(1);

}