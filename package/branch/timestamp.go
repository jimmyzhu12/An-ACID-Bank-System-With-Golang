package branch

import (
	"mp3/package/clog"
	comn "mp3/package/common"
	"time"
)

func (b *Branch) doTimestamp(command Command) {
	// need an if to see what command it is
	clog.DPrintf(clog.Blue, "the id of the command is %s", command.TransacId)
	//timestamp := b.transacs[command.TransacId].Timestamp
	timeStr := command.Timestamp
	timestamp, _ := time.Parse(time.RFC3339Nano, timeStr)
	_, exists := b.transacs[command.TransacId]
	if !exists {
		b.transacs[command.TransacId] = &Transaction{}
		b.transacs[command.TransacId].Timestamp = timestamp
	}
	account := command.Account
	branch := command.BranchId
	coordBranchId := command.CoordBranchId
	amount := command.Amount
	transacId := command.TransacId
	if branch != b.Id {
		clog.DPrintf(clog.Red, "this account is not in my branch")
	}

	if command.Type == DEPOSIT {
		_, exists := b.accounts[account]
		if !exists {
			impossibleTime := time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)
			// TODO: update time
			// impossibleTime := time.Time{}
			clog.DPrintf(clog.Blue, "Created new account %s ", account)
			b.accounts[account] = &Account{AccountName: account, CommitedValue: 0, CommitedTS: impossibleTime, RTS: make([]ReadTimestamp, 0), TW: make([]TentativeWrite, 0)}

			//TODO: change tentative write item naming
			tw := TentativeWrite{Timestamp: timestamp, Value: int(amount), Id: transacId}
			b.accounts[account].TW = append(b.accounts[account].TW, tw)
			clog.DPrintf(clog.Yellow, "the coordinator branch is %d", coordBranchId)
			res := CommandRes{TransacId: transacId, CoordBranchId: coordBranchId, Status: OK, Balance: 0, Type: DEPOSIT}
			clog.DPrintf(clog.Blue, "sending response: %s", res.Status)
			b.outCh <- res
		} else {
			b.timestampUpdateTWReply(transacId, account, amount, timestamp, coordBranchId)
		}
	} else if command.Type == WITHDRAW {
		_, exists := b.accounts[account]
		if !exists {
			// TODO: send back not found aborted
			res := CommandRes{TransacId: transacId, CoordBranchId: coordBranchId, Status: comn.NOT_FOUND, Balance: 0, Type: WITHDRAW}
			b.outCh <- res
			// broadcasting abort message to other
			for _, branch := range b.branches {
				abort := Command{
					Type:          ABORT,
					CoordBranchId: b.Id,
					BranchId:      branch.Id,
					Account:       command.Account,
					TransacId:     command.TransacId,
					Timestamp:     command.Timestamp,
				}
				if branch.Id == b.Id {
					b.doAbort(abort)
					continue
				}
				clog.DPrintf(clog.Blue, "Multicasint ABORT msg to %s", branch.Id.String())
				b.outCh <- abort
			}
		} else {
			amount *= -1
			b.timestampUpdateTWReply(transacId, account, amount, timestamp, coordBranchId)
		}
	} else if command.Type == BALANCE {
		_, exists := b.accounts[account]
		if !exists {
			clog.DPrintf(clog.Blue, "Aborted read")
			// TODO: send back new type not found aborted
			res := CommandRes{TransacId: transacId, CoordBranchId: coordBranchId, Status: comn.NOT_FOUND, Balance: 0, Type: BALANCE}
			b.outCh <- res
			for _, branch := range b.branches {
				abort := Command{
					Type:          ABORT,
					CoordBranchId: b.Id,
					BranchId:      branch.Id,
					Account:       command.Account,
					TransacId:     command.TransacId,
					Timestamp:     command.Timestamp,
				}
				if branch.Id == b.Id {
					b.doAbort(abort)
					continue
				}
				clog.DPrintf(clog.Blue, "Multicasint ABORT msg to %s", branch.Id.String())
				b.outCh <- abort
			}
		} else {
			b.attempRead(timestamp, account, transacId, coordBranchId, command)
		}
	}

}

func findMostRecentTime(reads []ReadTimestamp) time.Time {
	impossibleTime := time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)
	if len(reads) == 0 {
		return impossibleTime
	}
	mostRecent := reads[0]
	for _, read := range reads {
		if read.Timestamp.After(mostRecent.Timestamp) {
			mostRecent = read
		}
	}
	return mostRecent.Timestamp
}

func (b *Branch) findMostRecentWriteWithConstraint(account string, timestamp time.Time) TentativeWrite {
	// TODO: updat time
	impossibleTime := time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)
	if len(b.accounts[account].TW) == 0 {
		return TentativeWrite{Timestamp: impossibleTime}
	}
	mostRecent := b.accounts[account].TW[0]
	for _, tentwrite := range b.accounts[account].TW {
		if tentwrite.Timestamp.Before(timestamp) || tentwrite.Timestamp.Equal(timestamp) {
			if tentwrite.Timestamp.After(mostRecent.Timestamp) {
				mostRecent = tentwrite
			}
		}
	}
	return mostRecent
}

func (b *Branch) findMostRecentWriteWithConstraint_notself(account string, timestamp time.Time) TentativeWrite {
	// TODO: updat time
	impossibleTime := time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)
	if len(b.accounts[account].TW) == 0 {
		return TentativeWrite{Timestamp: impossibleTime}
	}
	mostRecent := TentativeWrite{Timestamp: impossibleTime}
	for _, tentwrite := range b.accounts[account].TW {
		clog.DPrintf(clog.Yellow, "the time stamp i am looking at is "+tentwrite.Timestamp.Format(time.RFC3339Nano))
		if tentwrite.Timestamp.Before(timestamp) {
			if tentwrite.Timestamp.After(mostRecent.Timestamp) {
				mostRecent = tentwrite
			}
		}
	}
	return mostRecent
}

func (b *Branch) updateTW(transacId string, account string, amount int, timestamp time.Time) {
	for index, tentwrite := range b.accounts[account].TW {
		if tentwrite.Id == transacId {
			// update the existing Tc entry in the TW list
			new_write := TentativeWrite{Timestamp: timestamp, Value: tentwrite.Value + amount, Id: transacId}
			b.accounts[account].TW[index] = new_write
			return
		}
	}
	// add a new write to TW
	new_write := TentativeWrite{Timestamp: timestamp, Value: amount, Id: transacId}
	b.accounts[account].TW = append(b.accounts[account].TW, new_write)
}

func (b *Branch) timestampUpdateTWReply(transacId string, account string, amount int, timestamp time.Time, srcbranch BranchID) {
	mostRecent := findMostRecentTime(b.accounts[account].RTS)
	if (timestamp.After(mostRecent) || timestamp.Equal(mostRecent)) && timestamp.After(b.accounts[account].CommitedTS) {
		b.updateTW(transacId, account, int(amount), timestamp)
		clog.DPrintf(clog.Blue, "updating tentative write account %s, amout %d", account, amount)
		// send back positive reply
		// TODO: ignore unused field
		res := CommandRes{TransacId: transacId, CoordBranchId: srcbranch, Status: OK, Balance: 0, Type: DEPOSIT}
		clog.DPrintf(clog.Blue, "sending response: %s", res.Status)
		b.outCh <- res
	} else { // abort transaction
		// send back negative reply
		res := CommandRes{TransacId: transacId, CoordBranchId: srcbranch, Status: ABORTED, Balance: 0, Type: DEPOSIT}
		clog.DPrintf(clog.Blue, "sending response: %s", res.Status)
		for _, branch := range b.branches {
			abort := Command{
				Type:          ABORT,
				CoordBranchId: b.Id,
				BranchId:      branch.Id,
				Account:       "",
				TransacId:     transacId,
				Timestamp:     timestamp.Format(time.RFC3339Nano),
			}
			if branch.Id == b.Id {
				b.doAbort(abort)
				continue
			}
			clog.DPrintf(clog.Blue, "Multicasint ABORT msg to %s", branch.Id.String())
			b.outCh <- abort
		}
		b.outCh <- res
	}
}

func (b *Branch) attempRead(timestamp time.Time, account string, transacId string, srcbranch BranchID, c Command) {
	if timestamp.After(b.accounts[account].CommitedTS) {
		mostRecent := b.findMostRecentWriteWithConstraint(account, timestamp)
		if b.accounts[account].CommitedTS.After(mostRecent.Timestamp) {
			// if Ds is committed, read Ds and add Tc to RTS list
			clog.DPrintf(clog.Cyan, "Committed -----------------")
			res := CommandRes{TransacId: transacId, CoordBranchId: srcbranch, Status: OK, Balance: b.accounts[account].CommitedValue, Type: BALANCE, BranchId: b.Id, Account: account}
			b.outCh <- res
			r := ReadTimestamp{Timestamp: timestamp, Id : transacId}
			b.accounts[account].RTS = append(b.accounts[account].RTS, r)
		} else {
			if mostRecent.Id == transacId { // if Ds was written by Tc, simply read Ds
				clog.DPrintf(clog.Cyan, "Written by me ------------------")
				res := comn.CommandRes{TransacId: transacId, CoordBranchId: srcbranch, Status: OK, Balance: mostRecent.Value + b.accounts[account].CommitedValue, Type: BALANCE, BranchId: b.Id, Account: account}
				b.outCh <- res

			} else { //wait until the transaction that wrote Ds is committed or aborted, and reapply the read rule
				// add a pending queue to solve this question
				// when mostRecent.Id commits or aborts, It will call dotimestamp on this command pushed
				// if b.transacs[mostRecent.Id] == nil {
				// 	clog.DPrintf(clog.Red, "not found this transacs in my transacs")
				// }
				impossibleTime := time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)
				if mostRecent.Id == "" && b.accounts[account].CommitedTS.Equal(impossibleTime) {
					// this means no one is writing this object, and this object should actually be treated as not existed
					// TODO: send back not found aborted 
					res := CommandRes{TransacId: transacId, CoordBranchId: srcbranch, Status: comn.NOT_FOUND, Balance: 0, Type: BALANCE, BranchId: b.Id, Account: account}
					clog.DPrintf(clog.Blue, "sending response: %s", res.Status)
					b.outCh <- res
					for _, branch := range b.branches {
						abort := Command{
							Type:          ABORT,
							CoordBranchId: b.Id,
							BranchId:      branch.Id,
							Account:       c.Account,
							TransacId:     c.TransacId,
							Timestamp:     c.Timestamp,
						}
						if branch.Id == b.Id {
							b.doAbort(abort)
							continue
						}
						clog.DPrintf(clog.Blue, "Multicasint ABORT msg to %s", branch.Id.String())
						b.outCh <- abort
					}
					return
				}
				clog.DPrintf(clog.Blue, "mostrecent %s,  current %s", mostRecent.Id, transacId)
				b.transacs[mostRecent.Id].PendingReads = append(b.transacs[mostRecent.Id].PendingReads, c)
			}
		}
	} else {
		clog.DPrintf(clog.Cyan, "REAL ABORT IN READ -----------------")
		res := CommandRes{TransacId: transacId, CoordBranchId: srcbranch, Status: ABORTED, Balance: 0, Type: BALANCE, BranchId: b.Id, Account: account}
		clog.DPrintf(clog.Blue, "sending response: %s", res.Status)
		for _, branch := range b.branches {
			abort := Command{
				Type:          ABORT,
				CoordBranchId: b.Id,
				BranchId:      branch.Id,
				Account:       c.Account,
				TransacId:     c.TransacId,
				Timestamp:     c.Timestamp,
			}
			if branch.Id == b.Id {
				b.doAbort(abort)
				continue
			}
			clog.DPrintf(clog.Blue, "Multicasint ABORT msg to %s", branch.Id.String())
			b.outCh <- abort
		}
		b.outCh <- res
	}
}
