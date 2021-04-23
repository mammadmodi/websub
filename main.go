package main

import (
	"os"
	"webchannel/ws"
)

func main() {
	os.Exit(run())
}

func run() int {
	if err := ws.ServeWS(); err != nil {
		return 1
	}
	return 0
}
