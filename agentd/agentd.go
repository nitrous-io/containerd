package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
)

func main() {
	flSocket := flag.String("c", "/var/run/agentd.sock", "socket of the agentd server")
	flag.Parse()

	if *flSocket == "" {
		fmt.Println("Missing socket address")
		os.Exit(1)
	}

	log.Fatalf("ListenAndServer: %v", http.ListenAndServe(*flSocket, nil))
}
