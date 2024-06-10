package branch

import "mp3/package/clog"

type CommitPhase int

const (
	Pending CommitPhase = iota
	Prepare
	Commit
	Aborted
	CommitOK
)

type CommitState struct {
	phase       CommitPhase
	numRecvdRes int
}

func (cs *CommitState) incrementRecvdRes() {
	cs.numRecvdRes += 1
}

func (cs *CommitState) updatePhase(phase CommitPhase) {
	cs.phase = phase
	cs.numRecvdRes = 0
}

func (b *Branch) handlePendingPhase(msg interface{}, cs *CommitState) {
	clog.DPrintf(clog.Yellow, "Handling Pending Stage Message")
	switch v := msg.(type) {
	case Command:
		// TODO: make sure the type of command
		cmd := v
		cs.updatePhase(Prepare)
		for _, branch := range b.branches {
			clog.DPrintf(clog.Yellow, "Start sending prepare to %s", branch.Id)
			prepare := Command{
				Type:          PREPARE,
				CoordBranchId: b.Id,
				BranchId:      branch.Id,
				Account:       cmd.Account,
				TransacId:     cmd.TransacId,
				Timestamp:     cmd.Timestamp,
			}
			b.outCh <- prepare
			clog.DPrintf(clog.Yellow, "End sending prepare to %s", branch.Id)
		}
	default:
		clog.DPrintf(clog.Red, "Invalid message in Pending phase")
	}
}

func (b *Branch) handlePreparePhase(msg interface{}, cs *CommitState) {
	clog.DPrintf(clog.Yellow, "Handling Prepare Stage Message")
	switch v := msg.(type) {
	case CommandRes:
		res := v
		switch res.Status {
		case OK:
			// TODO: make sure the type of response
			cs.incrementRecvdRes()
			if cs.numRecvdRes == b.numBr {
				cs.updatePhase(Commit)
				for _, branch := range b.branches {
					commit := Command{
						Type:          COMMIT,
						CoordBranchId: b.Id,
						BranchId:      branch.Id,
						Account:       res.Account,
						TransacId:     res.TransacId,
						Timestamp:     res.Timestamp,
					}
					b.outCh <- commit
				}
			}
		case ABORTED:
			cs.updatePhase(Aborted)
			for _, branch := range b.branches {
				abort := Command{
					Type:          ABORT,
					CoordBranchId: b.Id,
					BranchId:      branch.Id,
					Account:       res.Account,
					TransacId:     res.TransacId,
					Timestamp:     res.Timestamp,
				}
				b.outCh <- abort
			}
			b.outCh <- CommandRes{
				TransacId:     res.TransacId,
				Timestamp:     res.Timestamp,
				Type:          ABORT,
				Status:        ABORTED,
				CoordBranchId: res.CoordBranchId,
				BranchId:      res.BranchId,
			}
		default:
			clog.DPrintf(clog.Red, "Invalid status in Prepare phase")
		}
	default:
		clog.DPrintf(clog.Red, "Invalid message in Prepare phase")
	}
}

func (b *Branch) handleCommitPhase(msg interface{}, cs *CommitState) {
	clog.DPrintf(clog.Yellow, "Handling Commit Stage Message")
	switch v := msg.(type) {
	case CommandRes:
		res := v
		switch res.Status {
		case OK:
			cs.incrementRecvdRes()
			// TODO: make sure the type of response
			if cs.numRecvdRes == b.numBr {
				clog.DPrintf(clog.Yellow, "Successfully Committed")
				cs.updatePhase(CommitOK)
				res := CommandRes{
					TransacId:     res.TransacId,
					Timestamp:     res.Timestamp,
					CoordBranchId: res.CoordBranchId,
					BranchId:      res.BranchId,
					Type:          COMMIT,
					Status:        COMMIT_OK,
				}
				b.outCh <- res
			}
		case ABORTED:
			cs.updatePhase(Aborted)
			for _, branch := range b.branches {
				abort := Command{
					Type:          ABORT,
					CoordBranchId: b.Id,
					BranchId:      branch.Id,
					Account:       res.Account,
					TransacId:     res.TransacId,
					Timestamp:     res.Timestamp,
				}
				b.outCh <- abort
			}
			b.outCh <- CommandRes{
				TransacId:     res.TransacId,
				Timestamp:     res.Timestamp,
				Type:          ABORT,
				Status:        ABORTED,
				CoordBranchId: res.CoordBranchId,
				BranchId:      res.BranchId,
			}
		default:
			clog.DPrintf(clog.Red, "Invalid status in Commit phase")
		}
	default:
		clog.DPrintf(clog.Red, "Invalid message in Commit phase")
	}
}

func (b *Branch) handle2PhaseCommit(transacId string) {
	// TODO: check ingest branch message
	cs := CommitState{phase: Pending, numRecvdRes: 0}
	for msg := range b.transacs[transacId].commitCh {
		switch cs.phase {
		case Pending:
			b.handlePendingPhase(msg, &cs)
		case Prepare:
			b.handlePreparePhase(msg, &cs)
		case Commit:
			b.handleCommitPhase(msg, &cs)
		case Aborted, CommitOK:
			return
		default:
			clog.DPrintf(clog.Red, "Invalid message in during Commit")
		}
	}
}
