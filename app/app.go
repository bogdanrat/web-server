package app

import (
	"fmt"
	"github.com/bogdanrat/web-server/cache"
	"github.com/bogdanrat/web-server/config"
	"github.com/bogdanrat/web-server/handler"
	"github.com/bogdanrat/web-server/repository/postgres"
	"github.com/bogdanrat/web-server/router"
	"log"
	"net/http"
)

func Init() error {
	config.ReadFlags()
	if err := config.ReadConfiguration(); err != nil {
		return err
	}

	postgresDB, err := postgres.NewRepository(config.AppConfig.DB)
	if err != nil {
		return fmt.Errorf("could not establish database connection: %s", err.Error())
	}
	log.Println("Database connection established.")

	redisCache, err := cache.NewRedis(config.AppConfig.Redis)
	if err != nil {
		return fmt.Errorf("could not establish cache connection: %s", err.Error())
	}
	log.Println("Cache connection established.")

	handler.InitRepository(postgresDB)
	handler.InitCache(redisCache)

	redisCache.Subscribe("self", cache.HandleAuthServiceMessages, config.AppConfig.Authentication.Channel)

	return nil
}

func Start() {
	server := &http.Server{
		Addr:    config.AppConfig.Server.ListenAddress,
		Handler: router.New(),
	}

	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
