package branch

import comn "mp3/package/common"

type PeerBranch = comn.Branch
type BranchID = comn.BranchID

const (
	A = comn.A
	B = comn.B
	C = comn.C
	D = comn.D
	E = comn.E
)

type Command = comn.Command
type CommandType = comn.CommandType

const (
	BEGIN    = comn.BEGIN
	DEPOSIT  = comn.DEPOSIT
	BALANCE  = comn.BALANCE
	WITHDRAW = comn.WITHDRAW
	COMMIT   = comn.COMMIT
	ABORT    = comn.ABORT
	PREPARE  = comn.PREPARE
)

type CommandRes = comn.CommandRes
type CommandStatus = comn.CommandStatus

const (
	OK        = comn.OK
	ABORTED   = comn.ABORTED
	COMMIT_OK = comn.COMMIT_OK
)
