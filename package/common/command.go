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

type CommandType int

const (
	BEGIN CommandType = iota
	DEPOSIT
	BALANCE
	WITHDRAW
	COMMIT
	ABORT
	PREPARE
)

func (c CommandType) String() string {
	return [...]string{"BEGIN", "DEPOSIT", "BALANCE", "WITHDRAW", "COMMIT", "ABORT", "PREPARE"}[c]
}

func (c CommandType) Proto() message.CommandType {
	switch c {
	case BEGIN:
		return message.CommandType_BEGIN
	case DEPOSIT:
		return message.CommandType_DEPOSIT
	case BALANCE:
		return message.CommandType_BALANCE
	case WITHDRAW:
		return message.CommandType_WITHDRAW
	case COMMIT:
		return message.CommandType_COMMIT
	case ABORT:
		return message.CommandType_ABORT
	case PREPARE:
		return message.CommandType_PREPARE
	default:
		panic("unknown CommandType") // Handle unknown command types robustly in production code
	}
}

type Command struct {
	Type          CommandType
	CoordBranchId BranchID
	BranchId      BranchID
	Account       string
	Amount        int
	TransacId     string
	Timestamp     string
}

func (cmd Command) GetAccountName() string {
	if len(cmd.Account) == 0 {
		return ""
	}
	return cmd.BranchId.String() + "." + cmd.Account
}

// sendCommand sends a Command message over a TCP connection.
func SendCommand(conn net.Conn, cmd Command) error {
	data, err := proto.Marshal(CommandToProto(cmd))
	if err != nil {
		return err
	}
	clog.DPrintf(clog.Purple, "Sent data: <%s>", data)
	if err := binary.Write(conn, binary.BigEndian, uint32(len(data))); err != nil {
		return err
	}
	_, err = conn.Write(data)
	return err
}

// recvCommand receives a Command message from a TCP connection.
func RecvCommand(conn net.Conn) (*Command, error) {
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

	var cmdMsg message.Command
	if err := proto.Unmarshal(data, &cmdMsg); err != nil {
		return nil, err
	}
	cmd := ProtoToCommand(&cmdMsg)
	return &cmd, nil
}
