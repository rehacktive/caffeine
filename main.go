package main

import (
	"flag"
	"log"
	"os"
	"os/signal"

	"github.com/rehacktive/caffeine/database"
	"github.com/rehacktive/caffeine/service"
)

func main() {
	addr := flag.String("ip_port", "127.0.0.1:8000", "ip:port to expose")
	flag.Parse()

	server := service.Server{
		Address: *addr,
	}
	go server.Init(&database.MemDatabase{})

	log.Println("caffeine server started at " + server.Address)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	<-stop

	log.Println("bye")
}
