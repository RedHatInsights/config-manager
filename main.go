package main

import (
	"config-manager/api"
	"config-manager/application"
	"config-manager/config"
	"config-manager/infrastructure/kafka"
	"config-manager/infrastructure/persistence"
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

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

	accountRepo := persistence.AccountRepository{DB: db}

	//cm := ConfigManager{Config: config}
	cmService := application.ConfigManagerService{AccountRepo: &accountRepo}
	cmController := api.ConfigManagerController{
		ConfigManagerService: cmService,
	}

	cmController.Init()
	go cmController.Run(*cmAddr)

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
