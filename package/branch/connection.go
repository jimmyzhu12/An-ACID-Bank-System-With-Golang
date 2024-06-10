package branch

import (
	"bufio"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	comn "mp3/package/common"
	"net"
	"os"
	"strings"
	"sync"
	"time"

	clog "mp3/package/clog"
)

func (b *Branch) loadConfig(configPath string) {
	file, err := os.Open(configPath)
	if err != nil {
		fmt.Println("error in reading in config file:", err)
		return
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		words := strings.Split(line, " ")
		branchId := comn.StrToBranchID(words[0])
		hostname := words[1]
		port := words[2]
		b.branches[branchId] = &PeerBranch{Id: branchId, Hostname: hostname, Port: port}

		if branchId == b.Id {
			b.Port = port
			b.Hostname = hostname
		}
	}
}

func (b *Branch) connecToSinglePeerBranch(pb *PeerBranch, wg *sync.WaitGroup) {
	defer wg.Done()
	connected := false
	for !connected {
		conn, err := net.Dial("tcp", pb.GetIPAddr())
		if err != nil {
			clog.DPrintf(clog.Yellow, "Cannot connect to branch %s, retrying after 1 second...\n", pb.Id.String())
			time.Sleep(20 * time.Millisecond)
			continue
		}
		pb.Conn = conn
		//clog.DPrintf(clog.Green, "connected")
		fmt.Fprintf(conn, "%s\n", b.Id.String())
		clog.DPrintf(clog.Green, "<Branch %s> Connected!", pb.Id.String())
		connected = true
	}
}

func (b *Branch) connectToAllPeerBranch() {
	if b.Id.String() == "A" {
		var wg sync.WaitGroup
		wg.Add(len(b.branches) - 1)
		clog.DPrintf(clog.Blue, "connectToAllPeerBranch")
		for _, branch := range b.branches {
			if branch.Id != b.Id {
				go b.connecToSinglePeerBranch(branch, &wg)
			}
		}

		wg.Wait()
	}
	if b.Id.String() == "B" {
		var wg sync.WaitGroup
		wg.Add(len(b.branches) - 2)
		clog.DPrintf(clog.Blue, "connectToAllPeerBranch")
		for _, branch := range b.branches {
			if branch.Id != b.Id && branch.Id.String() != "A" {
				go b.connecToSinglePeerBranch(branch, &wg)
			}
		}

		wg.Wait()
	}
	if b.Id.String() == "C" {
		var wg sync.WaitGroup
		wg.Add(len(b.branches) - 3)
		clog.DPrintf(clog.Blue, "connectToAllPeerBranch")
		for _, branch := range b.branches {
			if branch.Id != b.Id && branch.Id.String() != "A" && branch.Id.String() != "B" {
				go b.connecToSinglePeerBranch(branch, &wg)
			}
		}
		wg.Wait()
	}
	if b.Id.String() == "D" {
		var wg sync.WaitGroup
		wg.Add(len(b.branches) - 4)
		clog.DPrintf(clog.Blue, "connectToAllPeerBranch")
		for _, branch := range b.branches {
			if branch.Id != b.Id && branch.Id.String() != "A" && branch.Id.String() != "B" && branch.Id.String() != "C" {
				go b.connecToSinglePeerBranch(branch, &wg)
			}
		}
		wg.Wait()
	}
}

func (b *Branch) identifyBranchConnection(conn net.Conn) {
	clog.DPrintf(clog.Blue, "identifyBranchConnection")
	reader := bufio.NewReader(conn)
	clog.DPrintf(clog.White, "the connection to identify is at port %s", conn.RemoteAddr().String())
	buf, err := reader.ReadString('\n')
	buf = strings.TrimSpace(buf)
	if err != nil {
		clog.DPrintf(clog.Red, "Error reading: %s", err.Error())
		return
	}

	branchId := comn.StrToBranchID(buf)
	if _, ok := b.branches[branchId]; !ok {
		clog.DPrintf(clog.Red, "Cannot identify the connected branch!")
		os.Exit(1)
	}
	b.branches[branchId].Conn = conn
	clog.DPrintf(clog.Green, "<Branch %s> Connected\n", branchId.String())
	clog.DPrintf(clog.Blue, "identifyBranchConnection finished")
}

func (b *Branch) waitForPeerBranchToConnect() {
	clog.DPrintf(clog.Blue, "waitForPeerBranchToConnect")
	var wg sync.WaitGroup
	ln, err := net.Listen("tcp", ":"+b.Port)
	if err != nil {
		clog.DPrintf(clog.Red, "Error listening: %s", err)
		return
	}
	b.ln = ln
	clog.DPrintf(clog.Green, "Listening on port %s....\n", b.Port)
	if b.Id.String() == "B" {
		for i := 0; i < 1; i++ {
			conn, err := ln.Accept()
			//clog.DPrintf(clog.Blue, "In waitForPeerBranch, this is round %d", i)
			if err != nil {
				clog.DPrintf(clog.Red, "Error accepting connection: %s\n", err)
				continue
			}

			wg.Add(1)
			go func() {
				defer wg.Done()
				b.identifyBranchConnection(conn)
			}()
		}
		wg.Wait()
	}
	if b.Id.String() == "C" {
		for i := 0; i < 2; i++ {
			conn, err := ln.Accept()
			//clog.DPrintf(clog.Blue, "In waitForPeerBranch, this is round %d", i)
			if err != nil {
				clog.DPrintf(clog.Red, "Error accepting connection: %s\n", err)
				continue
			}

			wg.Add(1)
			go func() {
				defer wg.Done()
				b.identifyBranchConnection(conn)
			}()
		}
		wg.Wait()
	}
	if b.Id.String() == "D" {
		for i := 0; i < 3; i++ {
			conn, err := ln.Accept()
			//clog.DPrintf(clog.Blue, "In waitForPeerBranch, this is round %d", i)
			if err != nil {
				clog.DPrintf(clog.Red, "Error accepting connection: %s\n", err)
				continue
			}

			wg.Add(1)
			go func() {
				defer wg.Done()
				b.identifyBranchConnection(conn)
			}()
		}
		wg.Wait()
	}
	if b.Id.String() == "E" {
		for i := 0; i < 4; i++ {
			conn, err := ln.Accept()
			//clog.DPrintf(clog.Blue, "In waitForPeerBranch, this is round %d", i)
			if err != nil {
				clog.DPrintf(clog.Red, "Error accepting connection: %s\n", err)
				continue
			}

			wg.Add(1)
			go func() {
				defer wg.Done()
				b.identifyBranchConnection(conn)
			}()
		}
		wg.Wait()
	}
	clog.DPrintf(clog.Blue, "finished waitForPeerBranchToConnect")
}

func (b *Branch) handleAllPeerBranchConnections() {
	clog.DPrintf(clog.Blue, "handleAllPeerBranchConnections")
	for _, branch := range b.branches {
		if branch.Id != b.Id {
			go b.handleSinglePeerBranchConnection(branch)
		}
	}
}

func (b *Branch) handleSinglePeerBranchConnection(branch *PeerBranch) {
	clog.DPrintf(clog.Yellow, "handle single peer connection %s", branch.Id)
	clog.DPrintf(clog.White, "handleSinglePeerBranchConnection listening on the addr %s, I am at the addr %s", branch.Conn.RemoteAddr().String(), branch.Conn.LocalAddr().String())
	reader := bufio.NewReader(branch.Conn)
	for {
		var length uint32
		err := binary.Read(reader, binary.BigEndian, &length)
		if err != nil {
			if err != io.EOF {
				fmt.Printf("Failed to read message length: %v\n", err)
			}
			return
		}
		data := make([]byte, length)
		clog.DPrintf(clog.Yellow, "got some data in handleSinglePeerBranchConnection")
		_, err = io.ReadFull(reader, data)
		if err != nil {
			fmt.Printf("Failed to read message data: %v\n", err)
			return
		}

		clog.DPrintf(clog.Blue, "received the message from peer"+string(data))
		b.ingestBrCh <- data
	}
}

func (b *Branch) monitorClientConnection() {
	clog.DPrintf(clog.Green, "Listening on port %s....\n", b.Port)
	clog.DPrintf(clog.Yellow, "I am monitoring client connection")
	for {
		conn, err := b.ln.Accept()
		if err != nil {
			clog.DPrintf(clog.Red, "Error accepting connection: %s\n", err)
			continue
		}

		go b.handleClientConnection(conn)
	}
}

func (b *Branch) processClientInitMsg(msg string) (string, time.Time, error) {
	if strings.Count(msg, "#") != 1 {
		return "", time.Time{}, errors.New("the input should contain exactly one '#' character")
	}

	parts := strings.Split(msg, "#")

	if len(parts) != 2 {
		return "", time.Time{}, errors.New("the input should split into exactly two parts")
	}

	parsedTime, err := time.Parse(time.RFC3339Nano, parts[1])
	if err != nil {
		return "", time.Time{}, fmt.Errorf("error parsing time: %v", err)
	}

	return parts[0], parsedTime, nil
}

func (b *Branch) handleClientConnection(conn net.Conn) {
	clog.DPrintf(clog.Blue, "Handling client connection")
	reader := bufio.NewReader(conn)
	buf, err := reader.ReadString('\n')
	buf = strings.TrimSpace(buf)
	if err != nil {
		clog.DPrintf(clog.Red, "Error reading: %s", err.Error())
		return
	}

	transacId, timestamp, err := b.processClientInitMsg(buf)
	if err != nil {
		clog.DPrintf(clog.Red, "Invalid Client Init Message")
	}
	clog.DPrintf(clog.Blue, "client Id %s", transacId)
	fmt.Println(timestamp)
	b.transacs[transacId] = &Transaction{Id: transacId, Timestamp: timestamp, Conn: conn}
	b.transacs[transacId].resCh = make(chan CommandRes, 20)
	b.transacs[transacId].PendingCommits = make([]comn.Command, 0)
	b.transacs[transacId].PendingReads = make([]comn.Command, 0)
	b.transacs[transacId].commitCh = make(chan interface{}, 20)
	b.transacs[transacId].abortCh = make(chan Command, 20)
	clog.DPrintf(clog.Green, "<Client %s> Connected\n", transacId)

	go b.handleTransaction(transacId)
}
