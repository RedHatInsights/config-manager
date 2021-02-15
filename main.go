package main

import (
	"config-manager/api/controllers"
	"config-manager/config"
	"config-manager/infrastructure"
	"config-manager/infrastructure/kafka"
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/labstack/echo/v4"

	_ "github.com/lib/pq"
)

func main() {
	ctx := context.Background()
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)

	config := config.Get()

	container := infrastructure.Container{Config: config}

	// Temporary - need to relocate api spec (and remove need for identity header)
	spec, err := controllers.GetSwagger()
	if err != nil {
		panic(err)
	}
	server := container.Server()
	server.GET("/openapi.json", func(ctx echo.Context) error {
		return ctx.JSON(http.StatusOK, spec)
	})

	configManager := container.CMController()
	configManager.Routes()

	go configManager.Start("0.0.0.0:8081")

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
