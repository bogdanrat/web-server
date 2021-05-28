package app

import (
	"fmt"
	pb "github.com/bogdanrat/web-server/contracts/proto/auth_service"
	"github.com/bogdanrat/web-server/service/core/cache"
	"github.com/bogdanrat/web-server/service/core/config"
	"github.com/bogdanrat/web-server/service/core/repository/postgres"
	"github.com/bogdanrat/web-server/service/core/router"
	"google.golang.org/grpc"
	"log"
	"net/http"
)

const (
	address = "localhost:50051"
)

var (
	httpRouter http.Handler
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

	conn, err := initGRPCConnection()
	if err != nil {
		return err
	}

	httpRouter = router.New(postgresDB, redisCache, pb.NewAuthClient(conn))

	redisCache.Subscribe("self", cache.HandleAuthServiceMessages, config.AppConfig.Authentication.Channel)

	return nil
}

func Start() {
	server := &http.Server{
		Addr:    config.AppConfig.Server.ListenAddress,
		Handler: httpRouter,
	}

	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}

func initGRPCConnection() (*grpc.ClientConn, error) {
	conn, err := grpc.Dial(address,
		grpc.WithInsecure(),
	)
	if err != nil {
		return nil, fmt.Errorf("error connecting to server: %v", err)
	}

	return conn, nil
}