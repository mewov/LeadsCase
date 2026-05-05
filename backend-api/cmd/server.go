package main

import (
	"demo-server/internal/app"
	"demo-server/internal/httpserver"
	libs2 "demo-server/internal/libs"
	"demo-server/internal/storage"
	"log"

	"github.com/go-chi/chi"
)

func main() {
	config, err := libs2.CreateConfig()
	if err != nil {
		log.Fatal("Invalid Environment:", err)
	}

	serverLogger := libs2.MustCreateLogger(config.ServerLogPath)
	serverHttpLogger := libs2.MustCreateLogger(config.ServerHttpLogPath)
	serverPostgresLogger := libs2.MustCreateLogger(config.ServerPostgresLogPath)

	serverLogger.Info("logger started")
	serverHttpLogger.Info("logger started")
	serverPostgresLogger.Info("logger started")

	url := storage.CreateUrlPostgres(config)
	postgres, err := storage.CreateAndConnectPostgres(url, serverPostgresLogger)
	if err != nil {
		serverLogger.Error("connect to postgres", "addr", config.PostgresAddr, "error", err.Error())
		log.Fatal("[-] Connect to postgres:", err)
	}
	serverLogger.Info("connected to postgres", "addr", config.PostgresAddr)
	log.Println("[+] Connect to postgres")

	err = postgres.Migration()
	if err != nil {
		log.Fatal("[-] Migration failed:", err)
	}
	serverLogger.Info("migration succeeded", "addr", config.PostgresAddr)
	log.Println("[+] Migration succeeded")

	err = app.SeedDefaultAdmin(postgres, config)
	if err != nil {
		serverLogger.Error("seed default admin", "error", err.Error())
		log.Fatal("[-] Seed default admin:", err)
	}
	serverLogger.Info("seed default admin", "addr", config.PostgresAddr)
	log.Println("[+] Seed default admin")

	router := chi.NewRouter()
	httpserver.Register(router, postgres, config, serverHttpLogger)

	serverLogger.Info("http server started", "addr", config.ServerAddr)
	if err = httpserver.Listen(router, config); err != nil {
		serverLogger.Error("http listen", "addr", config.PostgresAddr, "error", err.Error())
	}

	postgres.Close()
}
