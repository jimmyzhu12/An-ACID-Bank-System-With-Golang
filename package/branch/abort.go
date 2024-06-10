package branch

import "mp3/package/clog"

func (b *Branch) handleAbort(transacId string) {
	clog.DPrintf(clog.Cyan, "Start handling abort for transaction %s", transacId)
	for cmd := range b.transacs[transacId].abortCh {
		clog.DPrintf(clog.Blue, "Handling Abort cmd for Transaction %s", cmd.TransacId)
		b.cmdCh <- cmd
		for _, branch := range b.branches {
			clog.DPrintf(clog.Blue, "Multicasint ABORT msg to %s", branch.Id.String())
			abort := Command{
				Type:          ABORT,
				CoordBranchId: b.Id,
				BranchId:      branch.Id,
				Account:       cmd.Account,
				TransacId:     cmd.TransacId,
				Timestamp:     cmd.Timestamp,
			}
			b.outCh <- abort
		}
		b.outCh <- CommandRes{
			TransacId:     cmd.TransacId,
			Timestamp:     cmd.Timestamp,
			Type:          ABORT,
			Status:        ABORTED,
			CoordBranchId: cmd.CoordBranchId,
			BranchId:      cmd.BranchId,
		}
	}
}
