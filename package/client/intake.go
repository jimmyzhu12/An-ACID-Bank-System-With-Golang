package client

import (
	"bufio"
	"os"
	"strconv"
	"strings"

	comn "mp3/package/common"
)

func (c *Client) readCommand() {
	//TODO: add timestamp
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Text()
		words := strings.Split(line, " ")
		var cmd Command

		if words[0] == "BEGIN" {
			cmd.Type = BEGIN
		}
		if words[0] == "DEPOSIT" {
			cmd.Type = DEPOSIT
			x := strings.Split(words[1], ".")
			cmd.BranchId = comn.StrToBranchID(x[0])
			cmd.Account = x[1]
			cmd.Amount, _ = strconv.Atoi(words[2])
		}
		if words[0] == "BALANCE" {
			cmd.Type = BALANCE
			x := strings.Split(words[1], ".")
			cmd.BranchId = comn.StrToBranchID(x[0])
			cmd.Account = x[1]
		}
		if words[0] == "WITHDRAW" {
			cmd.Type = WITHDRAW
			x := strings.Split(words[1], ".")
			cmd.BranchId = comn.StrToBranchID(x[0])
			cmd.Account = x[1]
			cmd.Amount, _ = strconv.Atoi(words[2])
		}
		if words[0] == "COMMIT" {
			cmd.Type = COMMIT
		}
		if words[0] == "ABORT" {
			cmd.Type = ABORT
		}
		cmd.Timestamp = c.Timestamp
		c.cmdCh <- cmd
	}
}
