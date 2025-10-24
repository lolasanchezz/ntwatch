#include "./customTypes.h"
#include "./ex_ppi.c"


int main(int argc, char *argv[]) {
  int amtPids;
  int socketNum;

  pid_t *allPidsArr = getPids(&amtPids);

  pidsSocketsCountInternal(&socketNum, &amtPids, allPidsArr);
  struct socketInfo *goSocketData =
      (struct socketInfo *)malloc(socketNum * sizeof(struct socketInfo));

  goSocketStructs(goSocketData, socketNum);

  // for testing purposes only
 // printSockets(goSocketData, socketNum);
  printf("%d", socketNum);
  return 0;
}