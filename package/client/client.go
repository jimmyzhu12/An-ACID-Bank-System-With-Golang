package client

import (
	"log"
	"mp3/package/clog"
	"sync"
	"time"
)

func (c *Client) Init(id string, configPath string) {
	log.SetFlags(log.Lmsgprefix)

	c.Id = id
	c.Timestamp = time.Now().Format(time.RFC3339Nano)
	c.cmdCh = make(chan Command, 20)
	c.branches = make(map[BranchID]*Branch)
	c.loadConfig(configPath)
}

func (c *Client) Start() {
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		c.readCommand()
	}()
	go func() {
		defer wg.Done()
		c.handleCommand()
	}()
	wg.Wait()
	clog.DPrintf(clog.Green, "Client Terminated")
}
