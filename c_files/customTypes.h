#include <stdlib.h>

#define MAXCOMLEN 16
#define LISTENING_IP_PREFIX "0"
#define LISTENING_IP_PREFIX2 ":"

const char LOCAL_IPS[][16] = {
    "10.",        // 10.0.0.0 – 10.255.255.255
    "172.16.",    // 172.16.0.0 – 172.31.255.255
    "172.17.",
    "172.18.",
    "172.19.",
    "172.20.",
    "172.21.",
    "172.22.",
    "172.23.",
    "172.24.",
    "172.25.",
    "172.26.",
    "172.27.",
    "172.28.",
    "172.29.",
    "172.30.",
    "172.31.",
    "192.168.",   // 192.168.0.0 – 192.168.255.255
    "127.",       // localhost
        "fe80:"

};
#define LOCAL_IPS_LEN (sizeof(LOCAL_IPS)/sizeof(LOCAL_IPS[0]))

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

typedef ipAddr minSocketInfoArr[];

struct tcpInfo {};
