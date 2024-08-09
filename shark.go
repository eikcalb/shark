package main

import (
	"log/slog"
	"os"

	"eikcalb.dev/shark/src/app"
)

func main() {
	config, err := app.LoadConfig("config.json")
	if err != nil {
		slog.Error("Failed to load config", "error", err)
		os.Exit(1)
	}

	app := app.NewApplication(config)
	err = app.Run()
	if err != nil {
		slog.Error("Failed to run application", "error", err)
		os.Exit(1)
	}
}
