package api

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/aritradeveops/porichoy/internal/config"
	"github.com/aritradeveops/porichoy/internal/core/service"
	"github.com/aritradeveops/porichoy/internal/persistence/db"
	"github.com/aritradeveops/porichoy/internal/persistence/repository"
	"github.com/aritradeveops/porichoy/internal/ports/httpd"
	"github.com/aritradeveops/porichoy/internal/ports/httpd/handlers"
	"github.com/aritradeveops/porichoy/internal/ports/httpd/ui"
	"github.com/aritradeveops/porichoy/pkg/resolver"
)

func Run() error {
	config, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %v", err)
	}
	rvr := resolver.NewResolverFactory()
	r, err := rvr.Auto(config.Database.URIResolver)
	connectionString, err := r.Resolve(config.Database.URIResolver)
	if err != nil {
		return err
	}
	db := db.NewPostgres(connectionString.(string))
	err = db.Connect()
	if err != nil {
		return fmt.Errorf("failed to connect to database: %v", err)
	}
	dbtx, err := db.Tx()
	if err != nil {
		return fmt.Errorf("failed to connect to database: %v", err)
	}

	repo := repository.New(dbtx)
	srv := service.New(config, repo)
	handlers := handlers.New(srv)
	ui := ui.New(config.UI.Template)
	httpServer := httpd.NewServer(config, handlers, ui)

	go func() {
		err := httpServer.Start()
		if err != nil {
			log.Fatalf("failed to run the http server: %v", err)
		}
	}()
	quitChan := make(chan os.Signal, 1)
	signal.Notify(quitChan, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	<-quitChan
	fmt.Println("shutting down gracefully")
	err = httpServer.Shutdown()
	if err != nil {
		return fmt.Errorf("failed to shutdown http server: %v", err)
	}
	err = db.Disconnect()
	if err != nil {
		return fmt.Errorf("failed to disconnect from database: %v", err)
	}
	return nil
}
