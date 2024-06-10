package common

import (
	"net"
	"time"
)

type BranchID int

const (
	A BranchID = iota
	B
	C
	D
	E
)

func (b BranchID) String() string {
	names := []string{"A", "B", "C", "D", "E"}
	if b < A || b > E {
		return "Unknown"
	}
	return names[b]
}

func StrToBranchID(s string) BranchID {
	var nameToBranchID = map[string]BranchID{
		"A": A,
		"B": B,
		"C": C,
		"D": D,
		"E": E,
	}

	if id, ok := nameToBranchID[s]; ok {
		return id
	}
	return -1
}

// this is the branch copy for client use only
type Branch struct {
	Id       BranchID
	Hostname string
	Port     string
	Conn     net.Conn
}

func (b Branch) GetIPAddr() string {
	if len(b.Hostname) == 0 || len(b.Port) == 0 {
		return ""
	}
	return b.Hostname + ":" + b.Port
}

// Remember, there is 5 kind of messages received in this port
// 1. The client transaction
// 2. The commands forward by the other coordinate servers
// 3. The reply messages I got as a coordinate server
// 4. The 2PC, prepare messages from other coordinate servers
// 5. The 2PC, confirm/Abort messages from other coordinate servers
// And we only have one listening thread to distinguish different messages

// when init the branches, all of the branch should have a connection to each other to send
// that is to say, they should all init 4 threads to listen to 4 peers
type ConnMap struct {
	CMap map[int]net.Conn
}

// this is for coordinator server to save the transaction
// Ideally, the coordinator server will open a thread for each transaction received
// Probably use a map[ID] to track all of them?
type Transaction struct {
	Id string // Id of the client/transaction
	// NOTE IMPORTANT! I have to have a list because it is easier to roll back with a list
	Commands []Command // this is the commands that the client should send

	ExecuteIndex int // this is the index that I am going to execute next, also this is the reply index indicating which command i should be replying

	Timestamp time.Time // this timestamp is set to time.Now() as soon as first command is reached

	// NOTE, Ideally the client should be responsible for tracking the reponses
	Conn_C net.Conn // this the Connection I will be sending back to with a response
	// e.g. when the client receives first response, it auto maps it to the firststring command
}

type TentativeWrite struct {
	WTS time.Time // tentative write time
	TWV int       // tentative write values
	Id  string    // Id of the client/transaction , this is in case we need to abort this tentative write
}

type ReadTimestamp struct {
	RT time.Time
	Id string // Id of the client/transaction, this is in case we need to tell the contact server, I am replying to this specific transaction
}

// Ideally, the server will have a map to store Accounts
// key is the name of the account
type Account struct {
	AccountName   string
	CommitedValue int

	CommitedTS time.Time //   this is the timestamp index t

	RTS []ReadTimestamp  // read timestamp
	TW  []TentativeWrite // write timestamp
}
