package app

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	pb "github.com/bogdanrat/web-server/contracts/proto/storage_service"
	"github.com/bogdanrat/web-server/service/monitor"
	"github.com/bogdanrat/web-server/service/storage/config"
	"github.com/bogdanrat/web-server/service/storage/handler"
	"github.com/bogdanrat/web-server/service/storage/persistence/store"
	"github.com/bogdanrat/web-server/service/storage/persistence/store/diskstore"
	"github.com/bogdanrat/web-server/service/storage/persistence/store/s3store"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
	"log"
	"net"
	"net/http"
)

var (
	listener   net.Listener
	grpcServer *grpc.Server
)

func Init() error {
	var err error

	// init app config
	config.ReadFlags()
	if err = config.ReadConfiguration(); err != nil {
		return err
	}

	if err = initAwsSession(config.AppConfig.AWS); err != nil {
		return err
	}
	log.Println("AWS Session initialized.")

	// init storage
	var storage store.Store
	switch config.AppConfig.StorageEngine {
	case "disk":
		storage = diskstore.New(config.AppConfig.DiskStorage.Path)
	case "s3":
		storage = s3store.New(config.AWSSession, config.AppConfig.AWS.S3)
	default:
		return fmt.Errorf("unknown storage engine %s", config.AppConfig.StorageEngine)
	}

	if err = storage.Init(); err != nil {
		return err
	}

	if config.AppConfig.Prometheus.Enabled {
		initPrometheus()
	}

	// init grpc
	listener, err = net.Listen("tcp", config.AppConfig.Service.Address)
	if err != nil {
		return fmt.Errorf("failed to listen: %v", err)
	}

	var serverOptions []grpc.ServerOption
	grpcServer = grpc.NewServer(serverOptions...)

	storageServer := handler.New(storage)
	pb.RegisterStorageServer(grpcServer, storageServer)

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
	return nil
}

func Start() {
	log.Printf("gRPC listening on: %s\n", config.AppConfig.Service.Address)
	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

func initPrometheus() {
	_ = monitor.Setup()
	log.Println("Monitoring enabled.")

	router := gin.Default()
	router.Use(cors.Default())

	router.Use(monitor.PrometheusMiddleware())
	router.GET(config.AppConfig.Prometheus.MetricsPath, gin.WrapH(promhttp.Handler()))

	server := &http.Server{
		Addr:    config.AppConfig.Server.ListenAddress,
		Handler: router,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil {
			log.Fatal(err)
		}
	}()
}
