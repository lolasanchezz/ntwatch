## ntwatch
a better version of netstat and nettop which 
1. has a cleaner tui
2. includes information on associated proccesses through libproc


the goal is to create a mix between nettop and wireshark, while keeping it simple as nettop and making it more for 
a security overview rather than in depth testing.
 --> the biggest challenge is both figuring out macos's undocumented libproc kernel api, and synchronizing the packets/updating socket table properly through bubbletea's 
     elm architecture.
v1
    does have
        - information about sockets
        - a clean(ish) tui with bubbletea
        - both packet and socket functionality
    does not have
        - live updating
        - visual representation of sockets connected to packets

![img](./pics_for_readme/v1_ntwatch.png)
    
[![Athena Award Badge](https://img.shields.io/endpoint?url=https%3A%2F%2Faward.athena.hackclub.com%2Fapi%2Fbadge)](https://award.athena.hackclub.com?utm_source=readme)