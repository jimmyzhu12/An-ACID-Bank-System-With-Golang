package client

import (
	"fmt"
	"mp3/package/clog"
	comn "mp3/package/common"
	"time"
)

// GetCommandResTermOutput generates terminal output based on the CommandRes struct.
func GetCommandResTermOutput(cmdRes CommandRes) string {
	clog.DPrintf(clog.Blue, "The cmdRes.Type is %d", cmdRes.Type)
	switch cmdRes.Type {
	case BEGIN:
		return "OK"
	case DEPOSIT, WITHDRAW:
		if cmdRes.Status == OK {
			return "OK"
		} else if cmdRes.Status == ABORTED {
			return "ABORTED"
		} else if cmdRes.Status == comn.NOT_FOUND {
			return "NOT FOUND, ABORTED"
		}
	case BALANCE:
		if cmdRes.Status == OK {
			return fmt.Sprintf("%s = %d", cmdRes.GetAccountName(), cmdRes.Balance)
		} else if cmdRes.Status == comn.NOT_FOUND {
			return "NOT FOUND, ABORTED"
		} else if cmdRes.Status == ABORTED {
			return "ABORTED"
		}
	case COMMIT:
		if cmdRes.Status == COMMIT_OK {
			return "COMMIT OK"
		} else {
			return "ABORTED"
		}
	case ABORT:
		if cmdRes.Status == ABORTED {
			return "ABORTED"
		}
	}
	// Default case for unexpected or unhandled statuses
	return "UNEXPECTED STATUS"
}

func (c *Client) handleCommand() {
	for cmd := range c.cmdCh {
		clog.DPrintf(clog.Blue, "Get a command with tyep: %s", cmd.Type.String())
		conn := c.coordBr.Conn
		cmd.TransacId = c.Id
		cmd.CoordBranchId = c.coordBr.Id
		switch cmd.Type {
		case BEGIN:
			time.Sleep(100 * time.Millisecond)
			c.establishConnection()
			fmt.Println(GetCommandResTermOutput(CommandRes{Type: BEGIN}))
			time.Sleep(100 * time.Millisecond)
			continue
		default:
			comn.SendCommand(conn, cmd)
			clog.DPrintf(clog.Blue, "Sent transaction and waiting for response....")
		}

		cmdRes, err := comn.RecvCommandRes(conn)
		if err != nil {
			clog.DPrintf(clog.Red, "Error receiving command response")
		}
		clog.DPrintf(clog.Blue, "Received response: %s", cmdRes.Status)
		fmt.Println(GetCommandResTermOutput(*cmdRes))
		if cmdRes.Status == COMMIT_OK || cmdRes.Status == ABORTED || cmdRes.Status == comn.NOT_FOUND {
			return
		}
		//ime.Sleep(200 * time.Millisecond)
	}
}
