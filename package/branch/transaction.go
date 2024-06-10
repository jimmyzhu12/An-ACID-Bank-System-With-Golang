package branch

import (
	"io"
	"mp3/package/clog"
	comn "mp3/package/common"
	"net"
	"time"
)

type Transaction struct {
	Id             string
	Commands       []Command
	ExecuteIndex   int
	Timestamp      time.Time
	Conn           net.Conn
	resCh          chan CommandRes
	commitCh       chan interface{}
	abortCh        chan Command
	PendingReads   []Command
	PendingCommits []Command
}

func (b *Branch) handleTransaction(transacId string) {
	go b.handle2PhaseCommit(transacId)
	go b.handleAbort(transacId)

	t := b.transacs[transacId]
	clog.DPrintf(clog.Yellow, "handleTransaction transacId %s", transacId)
	for {
		cmd, err := comn.RecvCommand(t.Conn)
		if err != nil {
			if err != io.EOF {
				clog.DPrintf(clog.Red, "Error receiving command from client")
				continue
			}
			return
		}
		clog.DPrintf(clog.Yellow, "handleTransaction cmd %s", cmd.TransacId)
		b.ingestCliCh <- *cmd

		cmdRes := <-t.resCh
		clog.DPrintf(clog.Blue, "Sending to Client response fo type: %s", cmdRes.Type.String())
		comn.SendCommandRes(t.Conn, &cmdRes)
	}
}
