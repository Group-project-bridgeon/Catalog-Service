package main

import (
	"os"

	"github.com/Group-project-bridgeon/Catalog-Service/internal/app"
	"github.com/Group-project-bridgeon/Catalog-Service/internal/platform/logger"
)

func main() {
	log := logger.New("catalog")

	if err := app.Run(log); err != nil {
		log.Error().
			Err(err).
			Msg("catalog service failed")

		os.Exit(1)
	}
}
