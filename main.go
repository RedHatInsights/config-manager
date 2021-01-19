package main

import (
	"config-manager/config"
	"config-manager/infrastructure"
	"config-manager/infrastructure/kafka"
	"config-manager/utils"
	"context"
	"flag"
	"fmt"
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

	container := infrastructure.Container{Config: config}

	apiMux := container.Mux()
	apiSpec := container.ApiSpec()
	configManager := container.CMController()

	configManager.Routes()
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
