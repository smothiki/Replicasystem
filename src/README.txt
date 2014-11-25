I. SETUP

1. Install GO compiler
2. Set $GOPATH and $GOBIN
3. Put code into $GOPATH/src/github.com/replicasystem/
4. Put dependency go-simplejson in $GOPATH/src/github.com/bitly/go-simplejson

Example:
My $GOPATH is ~/gowork/, and server.go file can be found in 
~/gowork/src/github.com/replicasystem/src/server/server.go


============================================================

II. COMPILATION AND RUN

Run $GOPATH/src/github.com/replicasystem/src/launcher.sh [1-16]
This should start client, server and master with config01.json-config16.json

All logs are in $GOPATH/src/github.com/replicasystem/logs/,
where slog_X is log for servers, clog_X is for clients and mlog is for master

* X indicates the chain number of a bank (see below), so each file corresponds
  to a bank.


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
  This file generates a request list according to config file. It reads queries
  from a query file. If there are not enough queries, it generates queires
  randomly based on the probabilities read from config file.

utils.go (replicasystem/src/commons/utils/utils.go)
  Various utility functions that generates random UUIDs , utility functions that 
  read a value from Json file , implemented timeout utility function.

bank.go (replicasystem/src/commons/bank/bank.go)
  Structs related to banking operation, such as Transaction, Account and Bank,
  and relevant functions.

start.go (replicasystem/src/start/start.go) ***** deprecated *****
launcher (replicasystem/src/launcher)
end (replicasystem/src/end)
  launcher and start.go starts all clients, servers and master from one file. 
  end terminates all running processes. See Section II for details.

============================================================

IV. BUGS AND LIMITATIONS

No bugs have been observed by the time of submission.

Limitations:
  1. Hard coded IP addresses for ease of demonstration
  2. All chain servers should finish startup befoer master server starts.
     Otherwise, master may not receive any message from some servers during
     the first round of healh checking, then the servers will be identified
     died.

============================================================

V. CONTRIBUTIONS

Sivaram Mothiki - Bank transfer operation
Yansong Wang    - message syncronization among servers and master

============================================================

VI. OTHER COMMENTS

******[UPDATED IN PHASE 4 BELOW]******

A bank is identified as source bank or destination bank in a transfer
transaction by determining whether it's the sender or receiver of the transfer.

We make the tail server of source bank act as a client. Specifically, when
a transfer request arrives at the tail of source bank along the source bank
chain, tail query from master the current head of destination bank and sends
the request to the destination head. The tail of source bank then waits for the
tail of destination tail to reply. On receiving the reply, source tail replies
client as usual.

We added several fields to structure Reply/Request.

1. destBank, destAccount
These two fields are used to perform different actions on differ servers by
identifying server addresses and account IDs.

2. Sender, Receiver
Initially, all the fields are set as the client address, and when a server
received the message from client, it syncronizes the message to successor.
When the message arrives at source tail, it sets Sender as source tail and
Receiver as destination head. When destination head is about to reply source
tail, it sets the sender itself and Receiver source tail. In this way, messages
status can be easily identified by the server.


Additionally, source tail sends ack to predecessor only if destination head
replies to it, so when source tail fails, new source tail can resend transfer
requests in its 'Sent' list.


******[INHERITED FROM PHASE 3 BELOW]******

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
* If there are multiple chains, starting from chain1series, port number of the
  first server if each chain is chain1series*1000+1, (chain1series+1)*1000+1,
  (chain1series+2)*1000+1 ...

Since designated ports are binded to the UDP socket through which clients send
requested to servers, servers cannot use the same ports to send UDP packets
to master. Therefore, the sockets used to send message to master server are
bounded to ports whose numbers are desginated port numbers plus 100. For
instance, if a server is designated to 127.0.0.1:4002, it sends online message
to master via port 4102.

Requests and replies share the same structure (defined as structs.Request),
since they share some common fields. This makes it easy to transfer data
between clients are servers and servers themselves, but there is a 
Request.String() function, which prints different fields according to the
parameter, which is determined based on need.

Additionally, we use UDP protocol for communcations between clients and servers,
but HTTP protocol for communcations between servers to simplify our work. Since
the underlying implementation of HTTP is TCP, this does not incur any limitations.

Explanation of configure file. Take config01.json for example.

{
  "chains" : "1",                       //Number of chains
  "chainlength": "4",                   //Length of each chain
  "chain1series": "4",                  //Port numbers of first chain is 4XXX
  "MaxRequests":"20",                   //Each client send 20 queries, including
                                        //  those defined in config/request01.json
                                        //  and random ones generated based on
                                        //  parameters in config/request01.json
  "msgLossProb": "0.0",                 //  message loss probability
  "testrequests" : ["request01.json"],  //Quries are defined in config/request01.json
  "master" : "127.0.0.1:65535",         //Master address
  "checkOnlineCycle" : "3000",          //Master checks servers' status every 3000ms
  "sendOnlineCycle" : "1000",           //Servers send online(health) message every 1000ms
  "requestTimeout" : "5000",            //Client times out if does not receive
                                        //  response in 5000ms, not in use now
  "ackProcMaxTime" : "3000",            //Simulated ack processing max time
  "rqstProcMaxTime" : "1500",           //Simulated request process max time
  "extendSendInterval" : "500",         //Old tail sends bank information every
                                        //  500ms to new tail during extension
  "clientNum" : "1",                    //Number of clients
  "startDelay" : [0, 0, 0, 0],          //Start delay of each server in seconds,
                                        //  each element corresponds to a
                                        //  server, same below
  "lifetime" : [0, 11, 0, 0],           //Life time of each server
  "failOnReqSent" : [0, 0, 0, 0],       //Whether a server fails on determining
                                        //  what should be sent in 'Sent' to 
                                        //  successor. 1 for yes, 0 for no, same below
  "failOnRecvSent" : [0, 0, 0, 0],      //Whether a server fails on receiving
                                        //  'Sent' from predecessor
  "failOnExtension" : [0, 0, 0, 0]      //Whether a new server fails during extension
}

