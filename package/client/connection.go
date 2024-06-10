package client

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"math/rand"
	"time"
	"mp3/package/clog"
	comn "mp3/package/common"
)

func (c *Client) loadConfig(configPath string) {
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
		branch := comn.StrToBranchID(words[0])
		hostname := words[1]
		port := words[2]
		c.branches[branch] = &Branch{Id: branch, Hostname: hostname, Port: port}
		// conn will be updated later once we have a connection
	}
}

func GetCoordBranchID() BranchID {
	rand.Seed(time.Now().UnixNano())
	return BranchID(rand.Intn(5))

	//TODO: switch to the code above to randomly pick a coord branch
	//return 1
}

func (c *Client) establishConnection() {
	c.coordBr = *c.branches[GetCoordBranchID()]
	conn, err := net.Dial("tcp", c.coordBr.GetIPAddr())
	if err != nil {
		log.Fatal(err)
	}
	c.coordBr.Conn = conn
	fmt.Fprintf(conn, "%s#%s\n", c.Id, c.Timestamp)
	clog.DPrintf(clog.Green, "Established connection with Coord Branch %s!", c.coordBr.Id.String())
}
