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

	eventEmitter, eventListener, err := initMessageBroker(config.AppConfig.MessageBroker)
	log.Printf("Message Broker %s initialized.\n", config.AppConfig.MessageBroker.Broker)

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

func initMessageBroker(brokerConfig config.MessageBrokerConfig) (eventEmitter queue.EventEmitter, eventListener queue.EventListener, err error) {
	switch brokerConfig.Broker {
	case config.RabbitMQBroker:
		var conn *amqp.Connection
		amqpUri := fmt.Sprintf("amqp://%s:%s@%s:%s", brokerConfig.RabbitMQ.DefaultUser, brokerConfig.RabbitMQ.DefaultPassword, brokerConfig.RabbitMQ.Host, brokerConfig.RabbitMQ.Port)

		conn, err = amqp.Dial(amqpUri)
		if err != nil {
			err = fmt.Errorf("could not dial %s: %s", amqpUri, err)
			return
		}

		eventEmitter, err = amqp_queue.NewEventEmitter(conn, brokerConfig.RabbitMQ.Exchange)
		if err != nil {
			return
		}
		eventListener, err = amqp_queue.NewListener(conn, brokerConfig.RabbitMQ.Exchange, brokerConfig.RabbitMQ.Queue)
		if err != nil {
			return
		}

	default:
		err = fmt.Errorf("unknown message broker %s", brokerConfig.Broker)
	}

	return
}
