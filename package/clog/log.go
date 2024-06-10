package clog

import (
	"fmt"
	"log"
	"os"
	"time"
)

// func GenerateId() string {
// 	return uuid.New().String()
// }

// Debugging
const Debug = 1

func DPrintf(color Color, format string, a ...interface{}) (n int, err error) {
	log.SetPrefix(time.Now().Format("15:04:05.000 "))
	preset := getColorPreset(color)
	if Debug > 0 {
		str := fmt.Sprintf(format, a...)
		log.Print(preset + str + reset)
		os.Stderr.Sync()
	}
	return
}
