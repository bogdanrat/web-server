package app

import (
	"fmt"
	pb "github.com/bogdanrat/web-server/contracts/proto/auth_service"
	"github.com/bogdanrat/web-server/contracts/proto/database_service"
	"github.com/bogdanrat/web-server/contracts/proto/storage_service"
	"github.com/bogdanrat/web-server/service/core/cache"
	"github.com/bogdanrat/web-server/service/core/config"
	"github.com/bogdanrat/web-server/service/core/listener"
	"github.com/bogdanrat/web-server/service/core/mail"
	"github.com/bogdanrat/web-server/service/core/render"
	"github.com/bogdanrat/web-server/service/core/repository/postgres"
	"github.com/bogdanrat/web-server/service/core/router"
	"github.com/bogdanrat/web-server/service/queue"
	amqp_queue "github.com/bogdanrat/web-server/service/queue/amqp"
	"github.com/streadway/amqp"
	"google.golang.org/grpc"
	"log"
	"net/http"
)

var (
	httpRouter http.Handler
)

func Init() error {
	config.ReadFlags()
	if err := config.ReadConfiguration(); err != nil {
		return err
	}

	templateCache, err := render.CreateTemplateCache()
	if err != nil {
		return err
	}
	config.AppConfig.TemplateCache = templateCache

	if err = mail.NewService(config.AppConfig.SMTP); err != nil {
		return err
	}
	log.Println("SMTP Service initialized.")

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

	conn, err := initGRPCConnection(config.AppConfig.Services.Auth.GRPC.Address)
	if err != nil {
		return err
	}
	log.Printf("GRPC Dial %s successful.", config.AppConfig.Services.Auth.GRPC.Address)
	authClient := pb.NewAuthClient(conn)

	conn, err = initGRPCConnection(config.AppConfig.Services.Storage.GRPC.Address)
	if err != nil {
		return err
	}
	log.Printf("GRPC Dial %s successful.", config.AppConfig.Services.Storage.GRPC.Address)
	storageClient := storage_service.NewStorageClient(conn)

	conn, err = initGRPCConnection(config.AppConfig.Services.Database.GRPC.Address)
	if err != nil {
		return err
	}
	log.Printf("GRPC Dial %s successful.", config.AppConfig.Services.Database.GRPC.Address)
	databaseClient := database_service.NewDatabaseClient(conn)

	amqpConnection, err := amqp.Dial("amqp://guest:guest@localhost:5672")
	if err != nil {
		return err
	}
	eventEmitter, err := initEventEmitter(amqpConnection)
	if err != nil {
		return err
	}
	log.Println("Event Emitter initialized.")

	eventListener, err := initEventListener(amqpConnection)
	if err != nil {
		return err
	}
	log.Println("Event Listener initialized.")

	processor := listener.NewEventProcessor(eventListener)
	go func() {
		err := processor.ProcessEvent()
		if err != nil {
			log.Println(err)
		}
	}()

	httpRouter = router.New(postgresDB, redisCache, authClient, storageClient, databaseClient, eventEmitter)

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

func initGRPCConnection(addr string) (*grpc.ClientConn, error) {
	conn, err := grpc.Dial(addr,
		grpc.WithInsecure(),
	)
	if err != nil {
		return nil, fmt.Errorf("error connecting to server: %v", err)
	}

	return conn, nil
}

func initEventEmitter(conn *amqp.Connection) (queue.EventEmitter, error) {
	eventEmitter, err := amqp_queue.NewEventEmitter(conn, "authentication")
	if err != nil {
		return nil, err
	}
	return eventEmitter, nil
}

func initEventListener(conn *amqp.Connection) (queue.EventListener, error) {
	eventListener, err := amqp_queue.NewListener(conn, "authentication", "auth_queue")
	if err != nil {
		return nil, err
	}
	return eventListener, nil
}
