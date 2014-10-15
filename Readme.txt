git hub repo : https://github.com/smothiki/Replicasystem  

  INSTRUCTIONS.

Note # project Developed in go and distalgo . Inorder to run the project you should install GO and Distlago .

1. Multiple servers can be started using the file Replicasystem/src/startserver
  --- go run startserver.go
2. Clients can be run using the file Replicasystem/src/client or there is a client binary in Replicasystem/src/client

3. distalgo file can be started using command
   python -m da -f Replicasystem/distalgo/clientserver.da


  MAIN FILES

bank.py(Replicasystem/distalgo/bank.py) and bank.go (Replicasystem/src/commons/bank/bank.py)
  These files contains the software desing implementaion of a basic bank class that handles transactions like getbalance ,deposit, withdraw
  and also identifies if there is already any inconsistency or processed request is being served

server.go (Replicasystem/src/server/server.go)
  This file includes the entire server mechanism handling requests from client sending sync requests to next server in chain and tail sends
  processed request to client

client.go (Replicasystem/src/server/client.go)
  This file includes the entire client mechanism that sends requests to Head of the chain and sends GET requests to Tail of the chain
  listens for sync messages from tail

clientserver.da
  This file implemented in distalgo which generates processes for servers and clients according to the configuration specified
  and starts the processes and accordingly which server their purpose

structs/structs.go (Replicasystem/src/commons/structs/structs.go) and distalgo/structs.py
  This file delas with various data structures used in the software design like request , chain and generates the request and chain objects

structs/request-config.go (Replicasystem/src/commons//structs/structs.go) and distalgo/structs.py
  This file deals with generating requests according to probability . It takes probability of a request type and adjusts the number of various
  requests according to the probabity and maxrequests = 20

utils.go (Replicasystem/src/commons/utils/utils.go)
  Various utility functions that generates random UUIDs , utility functions that read a value from Json file , implemented timeout utility function

startclient.go and startserver.go (Replicasystem/src/commons/$1/$1)
  Starts the chain of servers and clients and setups the banks according to the specified configuration

config.json (Replicasystem/config/config.json)
  Has configuration for various entities like chain length , series , clients , totalchains

request.json (Replicasystem/config/request.json)
  has configuration about type of requests from which requests gets generated



  BUGS AND LIMITATIONS.  a list of all known bugs in and limitations

  As of now for Phase-2 We haven't observed any bugs . There are soem limitations
  1. distalgo the timeout is implemented for the entire list of process requests not for individual processes
  2. As we have implemented every thing in HTTP clients wont see any request drops unlike UDP clients , planning to move to UDP for phase -3
  3. Assumptions are that everything is an integer like balance and probability
  4. Total number of requests that are to be generated are hard coded to 20 . Phase 3 we will implement to read from config
  5. startclient.go won't start multiple clients , But the code works for multiple servers But we have to start another client for a different
     chain by changing the hardcoded values . Evenually automate this process in phase 3


  LANGUAGE COMPARISON (phase 3 only).  comparison of your experience
  implementing the algorithm in the two languages.

  CONTRIBUTIONS.

  OTHER COMMENTS.
  1. Implemented Timeouts in both languages . Which was very interesting and fun to learn
  2. Implemented a program that start multiple server and client processes similar to distalgo which was very fun and lot of learning in GO .
  3. Used an opensource library simplejson to make reading of json files easy in GO .
