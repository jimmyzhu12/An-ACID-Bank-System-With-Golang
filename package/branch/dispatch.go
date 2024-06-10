package branch

import (
	"encoding/binary"
	"log"
	"mp3/package/clog"
	comn "mp3/package/common"
	"mp3/package/message"
	"net"

	"google.golang.org/protobuf/proto"
)

func (b *Branch) Dispatch() {
	for data := range b.outCh {
		switch v := data.(type) {
		case Command:
			// Handle Command
			cmd := v
			if cmd.BranchId == b.Id {
				b.cmdCh <- cmd
			} else {
				clog.DPrintf(clog.Yellow, "when I am sending the commmand, the coordinator is %d", cmd.CoordBranchId)
				b.brOutCh <- cmd
			}
		case CommandRes:
			// Handle CommandRes
			res := v
			if b.Id == res.CoordBranchId {
				if res.Status == COMMIT_OK {
					b.transacs[res.TransacId].resCh <- res
				} else if res.Type == COMMIT || res.Type == PREPARE {
					b.transacs[res.TransacId].commitCh <- res
				} else {
					b.transacs[res.TransacId].resCh <- res
				}
			} else {
				b.brOutCh <- res
			}
		default:
			// Handle unexpected type
			log.Printf("Received data of unexpected type: %T\n", v)
		}
	}
}

func (b *Branch) sendBranchMessage() {
	var wrapper *message.BranchMessage
	var conn net.Conn
	for data := range b.brOutCh {
		switch v := data.(type) {
		case Command:
			// Handle Command
			wrapper = &message.BranchMessage{
				Payload: &message.BranchMessage_Command{
					Command: comn.CommandToProto(v),
				},
			}
			conn = b.branches[v.BranchId].Conn
			clog.DPrintf(clog.Blue, "Sent command message! send to %s, the connection is %d", v.BranchId, conn)
		case CommandRes:
			// Handle CommandRes
			wrapper = &message.BranchMessage{
				Payload: &message.BranchMessage_CommandRes{
					CommandRes: comn.CommandResToProto(v),
				},
			}
			conn = b.branches[v.CoordBranchId].Conn
			clog.DPrintf(clog.Blue, "Sent res message! send to %s, the status is %s, transactId: %s", v.CoordBranchId, v.Status.String(), v.TransacId)
		default:
			// Handle unexpected type
			log.Printf("Received data of unexpected type: %T\n", v)
		}
		data, err := proto.Marshal(wrapper)
		if err != nil {
			log.Fatalf("Failed to marshal CommandRes: %v", err)
		}
		bytes := 0
		clog.DPrintf(clog.White, "sendBranchMessage sending to the addr %s, I am at the addr %s", conn.RemoteAddr().String(), conn.LocalAddr().String())
		if err := binary.Write(conn, binary.BigEndian, uint32(len(data))); err != nil {
			log.Fatalf("Failed to write command to connection: %v", err)
		}

		if n, err := conn.Write(data); err != nil {
			log.Fatalf("Failed to write command to connection: %v", err)
		} else {
			bytes = n
		}
		clog.DPrintf(clog.Blue, "Sent %d bytes to connection", bytes)

	}
}
