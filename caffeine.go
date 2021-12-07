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
	projectVersion = "1.3.0"

	MEMORY = "memory"
	PG     = "postgres"
	FS     = "fs"
	SQLITE = "sqlite"

	// env
	envHostPort    = "IP_PORT"
	envDbType      = "DB_TYPE"
	envPgHost      = "PG_HOST"
	envPgUser      = "PG_USER"
	envPgPass      = "PG_PASS"
	envDbPath      = "DB_PATH"
	envAuthEnabled = "AUTH_ENABLED"
)

func main() {
	var addr, dbType, pgHost, pgUser, pgPass, dbPath string
	var authEnabled bool
	flag.StringVar(&addr, envHostPort, ":8000", "ip:port to expose")
	flag.StringVar(&dbType, envDbType, MEMORY, "db type to use, options: memory | postgres | fs")
	flag.StringVar(&pgHost, envPgHost, "0.0.0.0", "postgres host (port is 5432)")
	flag.StringVar(&pgUser, envPgUser, "", "postgres user")
	flag.StringVar(&pgPass, envPgPass, "", "postgres password")
	flag.StringVar(&dbPath, envDbPath, "./data", "path of the file storage root or sqlite")
	flag.BoolVar(&authEnabled, envAuthEnabled, false, "enable JWT auth")
	flag.Parse()

	server := service.Server{
		Address:     addr,
		AuthEnabled: authEnabled,
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
	case FS:
		db = &database.StorageDatabase{
			RootDirPath: dbPath,
		}
	case SQLITE:
		db = &database.SQLiteDatabase{
			DirPath: dbPath,
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
