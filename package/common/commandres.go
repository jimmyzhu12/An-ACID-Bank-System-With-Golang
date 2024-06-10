package common

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"mp3/package/clog"
	"mp3/package/message"
	"net"

	"google.golang.org/protobuf/proto"
)

type CommandStatus int

const (
	OK CommandStatus = iota
	ABORTED
	COMMIT_OK
	NOT_FOUND 
)

func (c CommandStatus) String() string {
	return [...]string{"OK", "ABORTED", "COMMIT_OK", "NOT_FOUND"}[c]
}

// Proto maps CommandStatus to its protobuf counterpart.
func (c CommandStatus) Proto() message.CommandStatus {
	switch c {
	case OK:
		return message.CommandStatus_OK
	case ABORTED:
		return message.CommandStatus_ABORTED
	case COMMIT_OK:
		return message.CommandStatus_COMMIT_OK
	case NOT_FOUND:
		return message.CommandStatus_NOT_FOUND
	default:
		panic("unknown CommandStatus") // Handle unknown status robustly in production code
	}
}

type CommandRes struct {
	TransacId     string
	Timestamp     string
	CoordBranchId BranchID
	BranchId      BranchID
	Account       string
	Type          CommandType
	Status        CommandStatus
	Balance       int
}

func (cmdRes CommandRes) GetAccountName() string {
	if len(cmdRes.Account) == 0 {
		return ""
	}
	return cmdRes.BranchId.String() + "." + cmdRes.Account
}

// sendCommandRes sends a CommandRes message over a TCP connection.
func SendCommandRes(conn net.Conn, cmdRes *CommandRes) error {
	data, err := proto.Marshal(CommandResToProto(*cmdRes))
	if err != nil {
		return err
	}
	if err := binary.Write(conn, binary.BigEndian, uint32(len(data))); err != nil {
		return err
	}
	_, err = conn.Write(data)
	return err
}

// recvCommandRes receives a CommandRes message from a TCP connection.
func RecvCommandRes(conn net.Conn) (*CommandRes, error) {
	var length uint32
	reader := bufio.NewReader(conn)
	err := binary.Read(reader, binary.BigEndian, &length)
	if err != nil {
		if err != io.EOF {
			fmt.Printf("Failed to read message length: %v\n", err)
			return nil, err
		}
		return nil, err
	}
	data := make([]byte, length)
	_, err = io.ReadFull(reader, data)
	if err != nil {
		fmt.Printf("Failed to read message data: %v\n", err)
		return nil, err
	}
	clog.DPrintf(clog.Purple, "Reveived data: <%s>", data)

	var cmdResMsg message.CommandRes
	if err := proto.Unmarshal(data, &cmdResMsg); err != nil {
		return nil, err
	}
	cmdRes := ProtoToCommandRes(&cmdResMsg)
	return &cmdRes, nil
}
