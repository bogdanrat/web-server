package app

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sts"
	pb "github.com/bogdanrat/web-server/contracts/proto/auth_service"
	"github.com/bogdanrat/web-server/contracts/proto/storage_service"
	"github.com/bogdanrat/web-server/service/core/cache"
	"github.com/bogdanrat/web-server/service/core/config"
	"github.com/bogdanrat/web-server/service/core/i18n/kvtranslator"
	"github.com/bogdanrat/web-server/service/core/listener"
	"github.com/bogdanrat/web-server/service/core/mail"
	"github.com/bogdanrat/web-server/service/core/render"
	"github.com/bogdanrat/web-server/service/core/router"
	"github.com/bogdanrat/web-server/service/core/store/dynamo"
	"github.com/bogdanrat/web-server/service/core/store/postgres"
	"github.com/bogdanrat/web-server/service/queue"
	amqp_queue "github.com/bogdanrat/web-server/service/queue/amqp"
	sqs_queue "github.com/bogdanrat/web-server/service/queue/sqs"
	"github.com/streadway/amqp"
	"google.golang.org/grpc"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

var (
	httpRouter http.Handler
)

type GRPCServiceID int

const (
	AuthService GRPCServiceID = iota
	StorageService
)

func Init() error {
	var err error
	if err = initTempDir(); err != nil {
		return err
	}

	config.ReadFlags()
	if err = config.ReadConfiguration(); err != nil {
		return err
	}

	// init aws
	if err = initAwsSession(config.AppConfig.AWS); err != nil {
		return err
	}
	log.Println("AWS Session initialized.")

	templateCache, err := render.CreateTemplateCache()
	if err != nil {
		return err
	}
	config.AppConfig.TemplateCache = templateCache

	if err = mail.NewService(config.AppConfig.SMTP); err != nil {
		return err
	}
	log.Println("SMTP Service initialized.")

	postgresDB, err := postgres.NewRepository()
	if err != nil {
		return fmt.Errorf("could not establish database connection: %s", err.Error())
	}
	log.Println("Database connection established.")

	redisCache, err := cache.NewRedis(config.AppConfig.Redis)
	if err != nil {
		return fmt.Errorf("could not establish cache connection: %s", err.Error())
	}
	log.Println("Cache connection established.")

	conn, err := initGRPC(AuthService)
	if err != nil {
		return err
	}
	log.Println("Auth Service GRPC connection established.")
	authClient := pb.NewAuthClient(conn)

	conn, err = initGRPC(StorageService)
	if err != nil {
		return err
	}
	log.Println("Storage Service GRPC connection established.")
	storageClient := storage_service.NewStorageClient(conn)

	// init emitter & listener
	eventEmitter, eventListener, err := initMessageBroker(config.AppConfig.MessageBroker)
	if err != nil {
		return fmt.Errorf("could not initialize %s message broker: %s", config.AppConfig.MessageBroker.Broker, err)
	}
	log.Printf("Message Broker %s initialized.\n", config.AppConfig.MessageBroker.Broker)

	// init store
	keyValueStore, err := dynamo.NewStore(config.AppConfig.I18N.TableName)
	if err != nil {
		return err
	}
	// init i18n
	translator, err := kvtranslator.New(keyValueStore)
	if err != nil {
		return err
	}

	processor := listener.NewEventProcessor(eventListener, translator)
	go func() {
		err := processor.ProcessEvent()
		if err != nil {
			log.Println(err)
		}
	}()

	httpRouter = router.New(postgresDB, redisCache, keyValueStore, authClient, storageClient, eventEmitter)

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

func initGRPC(service GRPCServiceID) (conn *grpc.ClientConn, err error) {
	var host string
	var port string

	switch service {
	case AuthService:
		host = os.Getenv("AUTH_SERVICE_HOST")
		port = os.Getenv("AUTH_SERVICE_PORT")
	case StorageService:
		host = os.Getenv("STORAGE_SERVICE_HOST")
		port = os.Getenv("STORAGE_SERVICE_PORT")
	default:
		return nil, fmt.Errorf("unknown grpc service id: %d", service)
	}

	target := fmt.Sprintf("%s:%s", host, port)
	conn, err = grpc.Dial(target,
		grpc.WithInsecure(),
	)

	if err != nil {
		err = fmt.Errorf("error dialing %s: %v", target, err)
		return
	}
	return
}

func initMessageBroker(brokerConfig config.MessageBrokerConfig) (eventEmitter queue.EventEmitter, eventListener queue.EventListener, err error) {
	switch brokerConfig.Broker {
	case config.RabbitMQBroker:
		var conn *amqp.Connection
		amqpUri := os.Getenv("RABBITMQ_URL")
		if amqpUri == "" {
			log.Printf("missing env rabbitmq url, using config url")
			amqpUri = fmt.Sprintf("amqp://%s:%s@%s:%s", brokerConfig.RabbitMQ.DefaultUser, brokerConfig.RabbitMQ.DefaultPassword, brokerConfig.RabbitMQ.Host, brokerConfig.RabbitMQ.Port)
		}

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
	case config.SQSBroker:
		sqsConfig := sqs_queue.Config{
			QueueName:                 brokerConfig.SQS.QueueName,
			ContentBasedDeduplication: brokerConfig.SQS.ContentBasedDeduplication,
			DelaySeconds:              brokerConfig.SQS.DelaySeconds,
			MessageRetentionPeriod:    brokerConfig.SQS.MessageRetentionPeriod,
			MaxNumberOfMessages:       brokerConfig.SQS.MaxNumberOfMessages,
			VisibilityTimeout:         brokerConfig.SQS.VisibilityTimeout,
			WaitTimeSeconds:           brokerConfig.SQS.WaitTimeSeconds,
		}
		eventEmitter, err = sqs_queue.NewEventEmitter(config.AWSSession, sqsConfig)
		if err != nil {
			return
		}
		eventListener, err = sqs_queue.NewEventListener(config.AWSSession, sqsConfig)
		if err != nil {
			return
		}

	default:
		err = fmt.Errorf("unknown message broker %s", brokerConfig.Broker)
	}

	return
}

/*
	Error: cannot parse multipart form: open /tmp/multipart-2576925029: no such file or directory.
	Platform: Docker

	In case the multipart data does not fit in the specified memory size, the data is written to a temporary file on disk instead.
	This file is created with ioutil.TempFile, and then os.TempDir() is called which:
	"The directory is neither guaranteed to exist nor have accessible permissions."
	To fix this, we ensure the directory exists.
*/
func initTempDir() error {
	// make sure we have a working tempdir, because:
	// os.TempDir(): The directory is neither guaranteed to exist nor have accessible permissions.
	tempDir := os.TempDir()
	if err := os.MkdirAll(tempDir, 1777); err != nil {
		return fmt.Errorf("failed to create temporary directory %s: %s", tempDir, err)
	}
	tempFile, err := ioutil.TempFile("", "genericInit_")
	if err != nil {
		return fmt.Errorf("failed to create tempFile: %s", err)
	}
	_, err = fmt.Fprintf(tempFile, "Hello, World!")
	if err != nil {
		return fmt.Errorf("failed to write to tempFile: %s", err)
	}
	if err := tempFile.Close(); err != nil {
		return fmt.Errorf("failed to close tempFile: %s", err)
	}
	if err := os.Remove(tempFile.Name()); err != nil {
		return fmt.Errorf("failed to delete tempFile: %s", err)
	}
	log.Printf("Using temporary directory %s", tempDir)
	return nil
}

func initAwsSession(awsConfig config.AWSConfig) error {
	sess, err := session.NewSession(&aws.Config{
		Region:                        aws.String(awsConfig.Region),
		CredentialsChainVerboseErrors: aws.Bool(true),
	})
	if err != nil {
		return err
	}

	config.SetAWSSession(sess)

	stsService := sts.New(sess)
	output, err := stsService.GetCallerIdentity(&sts.GetCallerIdentityInput{})
	if err != nil {
		return err
	}
	log.Printf("Caller Identity: %s\n", *output.Arn)

	return nil
}
