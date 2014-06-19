package main

import (
	"flag"
	"log"
	"net"
	"net/http"
	"os"
	"syscall"

	"github.com/nitrous-io/containerd/server"
)

const (
	defaultAgent = "docker_lxc"
)

func main() {
	flSocket := flag.String("c", "/var/run/containerd.sock", "socket of the continerd server")
	flag.Parse()

	if *flSocket == "" {
		log.Fatal("Missing socket address")
	}

	if err := syscall.Unlink(*flSocket); err != nil && !os.IsNotExist(err) {
		log.Fatal(err)
	}

	l, err := net.Listen("unix", *flSocket)
	if err != nil {
		log.Fatal(err)
	}

	s := server.NewServer("docker_lxc")
	httpServ := http.Server{Addr: *flSocket, Handler: s.Router}
	log.Fatalf("ListenAndServer: %v", httpServ.Serve(l))
}
