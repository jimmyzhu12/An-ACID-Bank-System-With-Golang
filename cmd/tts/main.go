package main

import (
	"mp3/package/branch"
	"os"
)

func main() {
	id := os.Args[1]
	configPath := os.Args[2]

	var b branch.Branch
	b.Init(id, configPath)
	b.Start()
}
