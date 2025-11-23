package main

import (
	"log"

	"github.com/carlosealves2/video-ia/service-discover/internal/bootstrap"
	"github.com/carlosealves2/video-ia/service-discover/internal/config"
)

func main() {
	cfg, err := config.NewBuilder().WithEnv().Validate().Build()
	if err != nil {
		log.Fatal(err)
	}

	app := bootstrap.New(cfg).
		InitLogger().
		InitRepository().
		InitHandlers().
		InitRouter()

	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}
