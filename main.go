package main

import (
	"config-manager/api"
	"config-manager/application"
	"config-manager/config"
	"config-manager/infrastructure/kafka"
	"config-manager/infrastructure/persistence"
	"config-manager/utils"
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

func main() {
	ctx := context.Background()
	var cmAddr = flag.String("cmAddr", ":8080", "Hostname:port of the config-manager server")
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)

	config := config.Get()

	connectionString := fmt.Sprintf("user=%s password=%s dbname=%s sslmode=disable",
		config.GetString("DBUser"),
		config.GetString("DBPass"),
		config.GetString("DBName"))

	db, err := sql.Open("postgres", connectionString)
	if err != nil {
		log.Fatal(err)
	}

	apiMux := mux.NewRouter()

	apiSpec := api.ApiSpecServer{
		Router:       apiMux,
		SpecFileName: config.GetString("ApiSpecFile"),
	}

	accountRepo := persistence.AccountRepository{DB: db}
	runRepo := persistence.RunRepository{DB: db}
	playRepo := persistence.PlaybookArchiveRepository{DB: db}

	cmService := application.ConfigManagerService{
		AccountRepo:  &accountRepo,
		RunRepo:      &runRepo,
		PlaybookRepo: &playRepo,
	}
	cmController := api.ConfigManagerController{
		ConfigManagerService: cmService,
		Router:               apiMux,
	}

	cmController.Routes()
	apiSpec.Routes()
	go utils.StartHTTPServer(*cmAddr, "config-manager", apiMux)

	resultsConsumer := kafka.NewResultsConsumer(config)
	connectionConsumer := kafka.NewConnectionsConsumer(config)

	defer func() {
		fmt.Println("Shutting down consumers")
		err := resultsConsumer.Close()
		if err != nil {
			fmt.Println("error closing results consumer")
			return
		}
		err = connectionConsumer.Close()
		if err != nil {
			fmt.Println("error closing connection consumer")
			return
		}
	}()

	go func() {
		for {
			fmt.Println("Results consumer running")
			m, _ := resultsConsumer.ReadMessage(ctx)
			fmt.Println(m)
		}
	}()

	go func() {
		for {
			fmt.Println("Connections consumer running")
			m, _ := connectionConsumer.ReadMessage(ctx)
			fmt.Println(m)
		}
	}()

	<-sigChan
}
