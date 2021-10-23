package main

import (
	"flag"
	"log"
	"os"
	"os/signal"

	"github.com/rehacktive/mvb/database"
)

func main() {
	addr := flag.String("ip_port", "127.0.0.1:8000", "ip:port to expose")
	flag.Parse()

	server := Server{
		address: *addr,
	}
	go server.Init(&database.MemDatabase{})

	log.Println("mvb server started at " + server.address)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	<-stop

	log.Println("bye")
}
