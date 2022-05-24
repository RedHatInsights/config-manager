package main

import (
	"config-manager/cmd"

	"github.com/rs/zerolog/log"
)

func main() {
	if err := cmd.Execute(); err != nil {
		log.Fatal().Err(err)
	}
}
