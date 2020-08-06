package main

import (
	"fmt"
	"net/http"
	"os"

	_ "github.com/joho/godotenv/autoload"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	api "github.com/dmitryt/image-previewer/internal/api"
	utils "github.com/dmitryt/image-previewer/internal/utils"
)

func init() {
	var val interface{}
	var logLevel zerolog.Level
	val, ok := os.LookupEnv("LOG_LEVEL")
	if ok {
		logLevel, _ = val.(zerolog.Level)
	} else {
		logLevel = zerolog.InfoLevel
	}
	zerolog.SetGlobalLevel(logLevel)
	fmt.Println("LLLL", zerolog.GlobalLevel())
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: zerolog.TimeFieldFormat})
}

func main() {
	address := fmt.Sprintf("0.0.0.0:%d", utils.Atoi(os.Getenv("PORT"), 8082))

	log.Info().Msgf("Server starting at %s", address)
	if err := http.ListenAndServe(address, api.Handlers()); err != nil {
		log.Fatal().Err(err).Msg("Startup failed")
	}
}
