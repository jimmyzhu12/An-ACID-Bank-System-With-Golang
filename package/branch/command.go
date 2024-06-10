package branch

import (
	"mp3/package/clog"
	comn "mp3/package/common"
	"time"
)

func (b *Branch) handleCommand() {
	for cmd := range b.cmdCh {
		clog.DPrintf(clog.Yellow, "handleCommand, the coo id is %d", cmd.CoordBranchId)
		if cmd.Type == COMMIT {
			b.doCommit(cmd)
		} else if cmd.Type == PREPARE {
			b.doPrepare(cmd)
		} else if cmd.Type == ABORT {
			b.doAbort(cmd)
		} else if cmd.Type == BALANCE || cmd.Type == DEPOSIT || cmd.Type == WITHDRAW {
			b.doTimestamp(cmd)
		}

	}
}

func (b *Branch) doPrepare(command Command) {
	// searching conflicts
	clog.DPrintf(clog.Yellow, "doPrepare on Branch %s", b.Id)
	clog.DPrintf(clog.Yellow, "preaparing the command %s", command.TransacId)
	impossibleTime := time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)
	max_closest_write := TentativeWrite{Timestamp: impossibleTime}
	for i, account := range b.accounts {
		clog.DPrintf(clog.Yellow, "search account %s", account.AccountName)
		myWrite := TentativeWrite{}
		writeFound := false
		for _, write := range account.TW {
			if write.Id == command.TransacId {
				myWrite = write
				writeFound = true
				clog.DPrintf(clog.Yellow, "found write in TW, timestamp is %s", write.Timestamp.Format(time.RFC3339Nano))
			}
		}
		if !writeFound {
			// no write, just ignore
			continue
		}
		// first need to check if I need to wait on any one
		closestWrite := b.findMostRecentWriteWithConstraint_notself(i, myWrite.Timestamp)
		clog.DPrintf(clog.Yellow, closestWrite.Timestamp.Format(time.RFC3339Nano))
		if closestWrite.Timestamp.After(impossibleTime) && !closestWrite.Timestamp.Equal(myWrite.Timestamp) {
			// this means we have a dependency
			clog.DPrintf(clog.Yellow, "find cloestWrite %s", closestWrite.Id)
			if max_closest_write.Timestamp.Before(closestWrite.Timestamp) {
				max_closest_write = closestWrite
			}
		}
	}
	if !max_closest_write.Timestamp.Equal(impossibleTime) {
		clog.DPrintf(clog.Yellow, "waiting for previous write to commit")
		b.transacs[max_closest_write.Id].PendingCommits = append(b.transacs[max_closest_write.Id].PendingCommits, command)
		return
	}
	// if I reach here it means there is no current tentative writes before me
	for _, account := range b.accounts {
		myWrite := TentativeWrite{}
		writeFound := false
		for _, write := range account.TW {
			if write.Id == command.TransacId {
				myWrite = write
				writeFound = true
			}
		}
		if !writeFound {
			// no write, just ignore
			continue
		}
		if (myWrite.Value + account.CommitedValue) < 0 {
			// abort this transaction if the commited value will fail under 0
			// also notify the coordinate server/ client
			// the client should accordingly just drop the whole transaction
			res := comn.CommandRes{TransacId: command.TransacId, CoordBranchId: command.CoordBranchId, Status: ABORTED, Balance: 0, Type: PREPARE, BranchId: command.BranchId}
			b.outCh <- res
			return
		}
	}
	// if the value test passed, just send ok to the prepare command
	res := comn.CommandRes{TransacId: command.TransacId, CoordBranchId: command.CoordBranchId, Status: OK, Balance: 0, Type: PREPARE}
	b.outCh <- res
}

func (b *Branch) doCommit(command Command) {
	// updating commited values
	clog.DPrintf(clog.Yellow, "doCommit on Branch %s", b.Id)
	impossibleTime := time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)
	for i, account := range b.accounts {
		new_RTS := make([]ReadTimestamp, 0)
		new_TW := make([]TentativeWrite, 0)
		for _, read := range account.RTS {
			if read.Id != command.TransacId {
				new_RTS = append(new_RTS, read)
			}
		}
		myWrite := TentativeWrite{Timestamp: impossibleTime}
		for _, write := range account.TW {
			if write.Id != command.TransacId {
				new_TW = append(new_TW, write)
			} else {
				myWrite = write
			}
		}
		b.UpdateCommitedValue(new_RTS, new_TW, myWrite, i)
	}
	res := comn.CommandRes{TransacId: command.TransacId, CoordBranchId: command.CoordBranchId, Status: OK, Balance: 0, Type: COMMIT}
	b.outCh <- res
   	b.printAllBalance(command)
	// handling any dependencies on me
	if b.transacs[command.TransacId] != nil && len(b.transacs[command.TransacId].PendingReads) != 0 {
		for _, attemptread := range b.transacs[command.TransacId].PendingReads {
			b.doTimestamp(attemptread)
		}
		b.transacs[command.TransacId].PendingReads = make([]comn.Command, 0)
	}
	// handle any pending commit
	if b.transacs[command.TransacId] != nil && len(b.transacs[command.TransacId].PendingCommits) != 0 {
		for _, t := range b.transacs[command.TransacId].PendingCommits {
			b.doPrepare(t)
		}
		b.transacs[command.TransacId].PendingCommits = make([]comn.Command, 0)
	}
}

func (b *Branch) UpdateCommitedValue(new_RTS []ReadTimestamp, new_TW []TentativeWrite, myWrite TentativeWrite, i string) {
	impossibleTime := time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)
	b.accounts[i].RTS = new_RTS
	b.accounts[i].TW = new_TW
	if myWrite.Timestamp == impossibleTime {
		return
	}
	b.accounts[i].CommitedValue += myWrite.Value
	b.accounts[i].CommitedTS = myWrite.Timestamp
}

func (b *Branch) doAbort(command Command) {
	for i, account := range b.accounts {
		new_RTS := make([]ReadTimestamp, 0)
		new_TW := make([]TentativeWrite, 0)
		for _, read := range account.RTS {
			if read.Id != command.TransacId {
				new_RTS = append(new_RTS, read)
			}
		}
		for _, write := range account.TW {
			if write.Id != command.TransacId {
				new_TW = append(new_TW, write)
			}
		}
		b.accounts[i].TW = new_TW
		b.accounts[i].RTS = new_RTS
	}
	// if command.CoordBranchId == b.Id {
	// 	res := comn.CommandRes{TransacId: command.TransacId, CoordBranchId: command.CoordBranchId, Status: ABORTED, Balance: 0, Type: ABORT}
	// 	b.outCh <- res
	// }
	// res := comn.CommandRes{TransacId: command.TransacId, CoordBranchId: command.CoordBranchId, Status: ABORTED, Balance: 0, Type: ABORT}
	// b.outCh <- res
	// handle any pending read
	b.printAllBalance(command)
	clog.DPrintf(clog.Blue, "the command.TransacId is "+command.TransacId)
	if b.transacs[command.TransacId] != nil && len(b.transacs[command.TransacId].PendingReads) != 0 {
		for _, attempread := range b.transacs[command.TransacId].PendingReads {
			clog.DPrintf(clog.Blue, "handle timestamp of transaction "+command.TransacId)
			clog.DPrintf(clog.Blue, "the attemp read transacId is "+attempread.TransacId)
			b.doTimestamp(attempread)
		}
		b.transacs[command.TransacId].PendingReads = make([]comn.Command, 0)
	}

	// handle any pending commit
	if b.transacs[command.TransacId] != nil && len(b.transacs[command.TransacId].PendingCommits) != 0 {
		for _, t := range b.transacs[command.TransacId].PendingCommits {
			b.doPrepare(t)
		}
		b.transacs[command.TransacId].PendingCommits = make([]comn.Command, 0)
	}
}

func (b *Branch) printAllBalance(command Command) {
	impossibleTime := time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)
	clog.DPrintf(clog.Cyan, "print all balances when committing/aborting for transac %s", command.TransacId)
	for _, value := range b.accounts {
		if !value.CommitedTS.Equal(impossibleTime) {
			clog.DPrintf(clog.Cyan, "%s . %s = %d", b.Id.String(), value.AccountName, value.CommitedValue)
		}
	}
}