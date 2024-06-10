# An ACID Bank System 



## Build Instructions

There should already be 2 executables in the main folder of our gitlab. “client” is the binary for our client part that takes in the command, hands it to the coordinate server and waits for server responses. “server” is the binary for our server part that reads from the client, redirects commands, and processes transactions. The binaries are for linux ubuntu. You can also run “make” to get the binaries manually.  You may need to download packages for GO, and protobuf if manually doing so.


## Design of system
I used timestamp ordering in this mp, which is a lock free algorithm that maintains ACID. 
Message formatting is handled through protobuf. 
### Server
There are generally 2 stages. The first stage is used for the server to connect to each other. The second stage is that once stage 1 is completed, the server starts to take in commands from the client and process them.
Each server maintains three main data structures (maps actually): branches, accounts, transacs. The “branches” data structure stores all the information about its peer branches, so that the current branch knows how to forward/receive command messages. The “accounts” data structure stores all information about the accounts stored on this server. That is to say: Committed Value, Committed TimeStamp, Tentative writes, ReadTimeStamps. The “transacs” data structure is mainly used to record the timestamp of this transaction and any pending reads or commits waiting for the current transaction to finish. If there is something pending, pop it out from the internal queue and do it. 
### Client
There is only one stage for the client. For each line from stdin, It just forwards it to a random server selected as coordinator, and waits for a response from the coordinator before it starts sending the next line. 
Each client maintains two important data structures: branches, timestamp. Similar as above, the “branches” data structure is used to store all the information to forward/receive commands/command responses. The “timestamp” data structure is used to record the timestamp for this client/transaction.

### Messages transferred
There are 3 kinds of data structure used to communicate between each server/client.
There is a large wrapper “BranchMessage”, and within the wrapper, there are “Command” and “CommandRes”
The Command message data structure is used for a coordinator to redirect command to its peers. Thus, it contains fields like CommandType, CoordBranchId, BranchId, Account, TransactionId, and TimeStamp. 
The Command response data structure is used for a processing server, after finishing processing the command, to communicate with the client (of course delivered through Coordinate servers). Thus, it contains information on Transaction Id, CoordBranchId, Account/Balance, Type, and most importantly Status.
Notice that, a command, and response don’t necessarily have to use all fields given in the data structure. 

### Communication Channels
For better explanation on Communication channels, please take a look at the graph
The channels/ dispatchers control the flow of all the messages within a Branch
The image can be found in the repo