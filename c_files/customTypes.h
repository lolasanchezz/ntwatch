#include <stdlib.h>
#define MAXCOMLEN 16
#define LISTENING_IP_PREFIX "0"
#define LISTENING_IP_PREFIX2 ":"
#pragma once
extern const char LOCAL_IPS[][16];
extern const int LOCAL_IPS_LEN;


typedef char ipAddr[16];
typedef uint16_t port;
typedef char processName_t[2 * MAXCOMLEN];
typedef __int32_t pid_t;

enum socket_type { UDP = 0, TCP };

enum connection_type { IPV4 = 1, IPV6 };

struct socketInfo {
  pid_t pid;
  processName_t processName;
  enum socket_type socket_type;         // UDP for socket_ln, TCP for socket_tcp
  enum connection_type connection_type; // helps distinguish below ipAddr
  ipAddr destIPAddr;                    // when ipv4, remaining 8 bytes are 0s
  port sourcePort;
  port destPort;
  int local;
  int listening;
};

//two primary functions which should be called from go
void goSocketStructs(void *goSocketData, int socketNum);
void socketCount(int *numSocketsPtr);
