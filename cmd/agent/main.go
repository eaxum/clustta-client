package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"clustta/internal/bridge"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	bridge.Start()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	fmt.Println("\nShutting down...")
	bridge.Stop()
}
