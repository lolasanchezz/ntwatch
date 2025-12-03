#include "./customTypes.h"
#include "./ex_ppi.c"


int main(int argc, char *argv[]) {
  printf("hello\n");

  int amtPids;
  int socketNum;

  pid_t *allPidsArr = getPids(&amtPids);
    printf("got pids\n");

  pidsSocketsCountInternal(&socketNum, &amtPids, allPidsArr);
      printf("internal socket count finished\n");

  
  struct socketInfo *goSocketData =
      (struct socketInfo *)malloc(socketNum * sizeof(struct socketInfo));
       printf("allocated memory\n");
 goSocketStructs(goSocketData, &socketNum);

  // for testing purposes only
  printSockets(goSocketData, socketNum, 1);
  
  printf("%d", socketNum);
  return 0;
}