package common

import "mp3/package/message"

func CommandToProto(cmd Command) *message.Command {
	return &message.Command{
		Type:          cmd.Type.Proto(),
		CoordBranchId: cmd.CoordBranchId.String(),
		BranchId:      cmd.BranchId.String(),
		Account:       cmd.Account,
		Amount:        int32(cmd.Amount),
		TransacId:     cmd.TransacId,
		Timestamp:     cmd.Timestamp,
	}
}

func ProtoToCommand(cmdMsg *message.Command) Command {
	return Command{
		Type:          CommandType(cmdMsg.Type),
		CoordBranchId: StrToBranchID(cmdMsg.CoordBranchId),
		BranchId:      StrToBranchID(cmdMsg.BranchId),
		Account:       cmdMsg.Account,
		Amount:        int(cmdMsg.Amount),
		TransacId:     cmdMsg.TransacId,
		Timestamp:     cmdMsg.Timestamp,
	}
}

func CommandResToProto(cmdRes CommandRes) *message.CommandRes {
	return &message.CommandRes{
		TransacId:     cmdRes.TransacId,
		CoordBranchId: cmdRes.CoordBranchId.String(),
		BranchId:      cmdRes.BranchId.String(),
		Account:       cmdRes.Account,
		Type:          cmdRes.Type.Proto(),
		Status:        cmdRes.Status.Proto(),
		Balance:       int32(cmdRes.Balance),
		Timestamp:     cmdRes.Timestamp,
	}
}

func ProtoToCommandRes(cmdResMsg *message.CommandRes) CommandRes {
	return CommandRes{
		TransacId:     cmdResMsg.TransacId,
		CoordBranchId: StrToBranchID(cmdResMsg.CoordBranchId),
		BranchId:      StrToBranchID(cmdResMsg.BranchId),
		Account:       cmdResMsg.Account,
		Type:          CommandType(cmdResMsg.Type),
		Status:        CommandStatus(cmdResMsg.Status),
		Balance:       int(cmdResMsg.Balance),
		Timestamp:     cmdResMsg.Timestamp,
	}
}
