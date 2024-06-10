package main

import (
	"fmt"
	"os"

	"mp3/package/client"
)

func main() {
	if len(os.Args) != 3 {
		fmt.Println("Usage: program <Id> <configPath>")
		os.Exit(1)
	}

	id := os.Args[1]
	configPath := os.Args[2]

	var client client.Client
	client.Init(id, configPath)
	client.Start()
}
