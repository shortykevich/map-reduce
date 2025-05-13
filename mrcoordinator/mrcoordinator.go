package main

import (
	"fmt"
	"os"
	"time"

	"github.com/shortykevich/map-reduce/mr"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: mrcoordinator inputfiles...\n")
		os.Exit(1)
	}

	m := mr.MakeCoordinator(os.Args[1:], 10)
	for m.Done() == false {
		time.Sleep(time.Second)
	}

	fmt.Println("Coordinator shutting down...")
	time.Sleep(time.Second)
}
