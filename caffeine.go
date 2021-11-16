package main

import (
	"log"
	"os"
	"os/signal"

	"github.com/namsral/flag"

	"github.com/rehacktive/caffeine/database"
	"github.com/rehacktive/caffeine/service"
)

const (
	projectName = (`
	 ██████╗ █████╗ ███████╗███████╗███████╗██╗███╗   ██╗███████╗
	██╔════╝██╔══██╗██╔════╝██╔════╝██╔════╝██║████╗  ██║██╔════╝
	██║     ███████║█████╗  █████╗  █████╗  ██║██╔██╗ ██║█████╗  
	██║     ██╔══██║██╔══╝  ██╔══╝  ██╔══╝  ██║██║╚██╗██║██╔══╝  
	╚██████╗██║  ██║██║     ██║     ███████╗██║██║ ╚████║███████╗
	 ╚═════╝╚═╝  ╚═╝╚═╝     ╚═╝     ╚══════╝╚═╝╚═╝  ╚═══╝╚══════╝	
	`)
	projectVersion = "0.1"

	MEMORY = "memory"
	PG     = "postgres"

	// env
	envHostPort = "IP_PORT"
	envDbType   = "DB_TYPE"
	envPgHost   = "PG_HOST"
	envPgUser   = "PG_USER"
	envPgPass   = "PG_PASS"
)

func main() {
	var addr, dbType, pgHost, pgUser, pgPass string
	flag.StringVar(&addr, envHostPort, ":8000", "ip:port to expose")
	flag.StringVar(&dbType, envDbType, MEMORY, "db type to use, options: memory | postgres")
	flag.StringVar(&pgHost, envPgHost, "0.0.0.0", "postgres host (port is 5432)")
	flag.StringVar(&pgUser, envPgUser, "", "postgres user")
	flag.StringVar(&pgPass, envPgPass, "", "postgres password")
	flag.Parse()

	server := service.Server{
		Address: addr,
	}

	var db service.Database
	switch dbType {
	case MEMORY:
		db = &database.MemDatabase{}
	case PG:
		db = &database.PGDatabase{
			Host: pgHost,
			User: pgUser,
			Pass: pgPass,
		}
	}
	go server.Init(db)

	log.Println(projectName, " version: ", projectVersion)
	log.Println("server started at " + server.Address)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	<-stop

	log.Println("bye")
}
