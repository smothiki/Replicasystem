I. SETUP

THIS SECTION IS FOR GO VERSION ONLY

1. Install GO compiler
2. Set $GOPATH and $GOBIN
3. Put code into $GOPATH/src/github.com/replicasystem/

Example:
My $GOPATH is ~/gowork/, and server.go file can be found in 
~/gowork/src/github.com/replicasystem/src/server/server.go


============================================================

II. COMPILATION AND RUN


A. GO VERSION

There are two ways to run the programs.

1. Make sure current directory is $GOPATH/src/github.com/replicasystem/src/
   Then execute the following instructions:

> go install client/client.go
> go install master/master.go
> go install server/server.go
> go run start/start.go config0[1-8].json

* config01.json - config08.json are test cases.
* This method does NOT print real-time stdout information.

2. Run $GOPATH/src/github.com/replicasystem/src/launcher.sh [1-8]
This should start client, server and master with config01.json-config08.json

* This method print real-time stdout information, and run client, servers
  and master in 3 separate terminals


All logs are in $GOPATH/src/github.com/replicasystem/logs/,
where slog is log for servers, clog is for clients and mlog is for master


B. DISTALGO VERSION


============================================================

III. MAIN FILES

server.go (replicasystem/src/server/server.go)
  This file includes the entire server mechanisms, including handling requests
  from clients, sending sync requests to successor, handling chain modification
  and extension and so on.

client.go (replicasystem/src/client/client.go)
  This file includes the entire client mechanisms that send requests to head or
  tail of the chain with UDP sockets.

structs/structs.go (replicasystem/src/commons/structs/structs.go) 
  This file deals with various data structures used in the software design like 
  request, chain and generates the request and chain objects.

structs/request-config.go (replicasystem/src/commons/structs/request-config.go) 
  This file deals with generating requests according to probability . It takes 
  probability of a request type and adjusts the number of various
  requests according to the probabity and maxrequests = 20

utils.go (replicasystem/src/commons/utils/utils.go)
  Various utility functions that generates random UUIDs , utility functions that 
  read a value from Json file , implemented timeout utility function.

bank.go (replicasystem/src/commons/bank/bank.go)
  Structs related to banking operation, such as Transaction, Account and Bank,
  and relevant functions.

start.go (replicasystem/src/start/start.go)
launcher.sh (replicasystem/src/launcher.sh)
  Starts all clients, servers and master from one file. See Section II for details.

============================================================

IV. BUGS AND LIMITATIONS

1. Hard coded IP addresses
2. Hard coded random requests generation
3. Balance / Amount int
4. Doesn't resend requests on time out

============================================================

V. LANGUAGE COMPARISON


============================================================

VI. CONTRIBUTIONS

============================================================

VII. OTHER COMMENTS

A. GO VERSION

For purpose of demonstration, we don't read IP addresses from config file.
Instead, they are all assigned to the IP address 127.0.0.1 with different
port numbers. This makes it easy to read and modify configuration files.
However, this can be easily modified to adapt to different IP addresses and
port numbers. The logics we use here are as follows.

* All servers read IP address (and port) of master server from config file.
* Starting from chain1series (in config file), based on the number of chains,
  servers are assgined to ports series*1000+1, series*1000+2, series*1000+3...
  For example, if there is only one chain, namely one bank, and chain1series=4,
  then servers are designated to ports 4001, 4002, 4003 ...
* Port numbers assigned to clients are series*1000+999, series*1000+998 ...
  For example, if chain1series=4 and there is only one client, then its
  port number is 4999

Since designated ports are binded to the UDP socket through which clients send
requested to servers, servers cannot use the same port to send UDP packets
to master. Therefore, the sockets used to send message to master server are
bounded to ports whose numbers are desginated port numbers plus 100.

Requests and replies share the same structure (defined as structs.Request),
since they share some common fields. This makes it easy to transfer data
between clients are servers and servers themselves, but there is a 
Request.String() function, which prints different fields according to the
parameter, which is determined based on need.

Additionally, we use HTTP protocol to simplify our work, but since the
underlying implementation of HTTP is TCP, this does not incur any limitations.
