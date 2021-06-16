package app

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sts"
	pb "github.com/bogdanrat/web-server/contracts/proto/database_service"
	"github.com/bogdanrat/web-server/service/database/config"
	"github.com/bogdanrat/web-server/service/database/db/postgres"
	"github.com/bogdanrat/web-server/service/database/handler"
	"google.golang.org/grpc"
	"log"
	"net"
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

	// init aws
	if err = initAwsSession(config.AppConfig.AWS); err != nil {
		return err
	}
	log.Println("AWS Session initialized.")

	// init database
	database, err := postgres.NewDatabase()
	if err != nil {
		return fmt.Errorf("could not establish database connection: %s", err.Error())
	}
	log.Println("Database connection established.")

	// init grpc
	listener, err = net.Listen("tcp", config.AppConfig.Service.Address)
	if err != nil {
		return fmt.Errorf("failed to listen: %v", err)
	}

	var serverOptions []grpc.ServerOption
	grpcServer = grpc.NewServer(serverOptions...)

	storageServer := handler.New(database)
	pb.RegisterDatabaseServer(grpcServer, storageServer)

	return nil
}

func Start() {
	log.Printf("gRPC listening on: %s\n", config.AppConfig.Service.Address)
	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
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
