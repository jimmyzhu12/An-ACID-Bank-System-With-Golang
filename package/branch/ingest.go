package branch

import (
	"log"
	"mp3/package/clog"
	comn "mp3/package/common"
	"mp3/package/message"

	"google.golang.org/protobuf/proto"
)

func (b *Branch) ingestBranchMessages() {
	clog.DPrintf(clog.Blue, "Ingesting branch messages")
	for msg := range b.ingestBrCh {
		clog.DPrintf(clog.Blue, "Ingesting msg: %s", string(msg))
		wrapper := &message.BranchMessage{}
		if err := proto.Unmarshal([]byte(msg), wrapper); err != nil {
			log.Fatalf("Failed to unmarshal response: %v", err)
		}

		switch x := wrapper.Payload.(type) {
		case *message.BranchMessage_Command:
			// Handle Command
			cmd := comn.ProtoToCommand(wrapper.GetCommand())
			if (cmd.Type == COMMIT || cmd.Type == PREPARE) && cmd.CoordBranchId == b.Id {
				if transac, ok := b.transacs[cmd.TransacId]; ok {
					transac.commitCh <- cmd
				} else {
					res := comn.CommandRes{
						TransacId:     cmd.TransacId,
						CoordBranchId: cmd.CoordBranchId,
						BranchId:      cmd.BranchId,
						Status:        OK,
						Type:          PREPARE,
					}
					b.outCh <- res
				}
			} else {
				b.cmdCh <- cmd
			}
			clog.DPrintf(clog.Blue, "Received Command: %v", x.Command)
		case *message.BranchMessage_CommandRes:
			// Handle CommandRes
			cmdRes := comn.ProtoToCommandRes(wrapper.GetCommandRes())
			if cmdRes.Type == COMMIT || cmdRes.Type == PREPARE {
				b.transacs[cmdRes.TransacId].commitCh <- cmdRes
			} else {
				b.outCh <- cmdRes
			}
			clog.DPrintf(clog.Blue, "Received Command Response: %v", x.CommandRes)
		default:
			log.Printf("Received an unknown type of message")
		}
	}
}

func (b *Branch) ingestClientMessages() {
	for cmd := range b.ingestCliCh {
		clog.DPrintf(clog.Blue, "Received Cli Msg with TransacId: %s", cmd.TransacId)
		if cmd.Type == COMMIT {
			b.transacs[cmd.TransacId].commitCh <- cmd
		} else if cmd.Type == ABORT {
			b.transacs[cmd.TransacId].abortCh <- cmd
		} else {
			b.outCh <- cmd
		}
	}
}
