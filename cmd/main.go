package main

import (
	"log"

	"github.com/yiannis54/go-socket-server/internal/app"
	"github.com/yiannis54/go-socket-server/internal/config"
)

func main() {
	cfg, err := config.LoadConfiguration()
	if err != nil {
		log.Fatal(err)
	}

	if err := app.Run(cfg); err != nil {
		log.Fatal(err)
	}
}
