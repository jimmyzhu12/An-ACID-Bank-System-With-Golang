package branch

import (
	"log"
	"mp3/package/clog"
	comn "mp3/package/common"
	"net"
	"sync"
)

var TotalBranchNum = 5

type Branch struct {
	Id          BranchID
	Hostname    string
	Port        string
	ln          net.Listener
	branches    map[BranchID]*PeerBranch
	numBr       int
	accounts    map[string]*Account
	transacs    map[string]*Transaction
	ingestBrCh  chan []byte
	ingestCliCh chan Command
	cmdCh       chan Command
	outCh       chan interface{}
	brOutCh     chan interface{}
}

func (b *Branch) Init(id string, configPath string) {
	log.SetFlags(log.Lmsgprefix)

	b.Id = comn.StrToBranchID(id)
	b.branches = make(map[BranchID]*PeerBranch)
	b.numBr = 5
	b.accounts = make(map[string]*Account)
	b.transacs = make(map[string]*Transaction)

	b.ingestBrCh = make(chan []byte, 20)
	b.ingestCliCh = make(chan Command, 20)
	b.cmdCh = make(chan Command, 20)
	b.outCh = make(chan interface{}, 20)
	b.brOutCh = make(chan interface{}, 20)

	b.loadConfig(configPath)
}

func (b *Branch) Start() {
	clog.DPrintf(clog.Blue, "starting branch")

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		b.connectToAllPeerBranch()
	}()

	go func() {
		defer wg.Done()
		b.waitForPeerBranchToConnect()
	}()

	wg.Wait()

	go b.handleAllPeerBranchConnections()
	//time.Sleep(1 * time.Second)
	// time.Sleep(1 * time.Second)
}

func (b *Branch) Run() {
	clog.DPrintf(clog.Blue, "Branch %s Running...", b.Id.String())
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		b.monitorClientConnection()
	}()
	go func() {
		defer wg.Done()
		b.ingestBranchMessages()
	}()
	go func() {
		defer wg.Done()
		b.ingestClientMessages()
	}()
	go func() {
		defer wg.Done()
		b.Dispatch()
	}()
	go func() {
		defer wg.Done()
		b.sendBranchMessage()
	}()
	go func() {
		defer wg.Done()
		b.handleCommand()
	}()
	wg.Wait()
	clog.DPrintf(clog.Blue, "Branch %s Closed.", b.Id.String())
}
