package app

import (
	"fmt"
	"github.com/bogdanrat/web-server/service/auth/config"
	"github.com/bogdanrat/web-server/service/auth/handler"
	"github.com/bogdanrat/web-server/service/auth/interceptor"
	pb "github.com/bogdanrat/web-server/service/auth/proto"
	"go.opencensus.io/examples/exporter"
	"go.opencensus.io/plugin/ocgrpc"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/zpages"
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

	config.ReadFlags()
	if err = config.ReadConfiguration(); err != nil {
		return err
	}

	serverOptions := []grpc.ServerOption{
		grpc.UnaryInterceptor(interceptor.RequestDurationInterceptor),
	}

	if config.AppConfig.OpenCensus.Enabled {
		initOpenCensus()
		serverOptions = append(serverOptions, grpc.StatsHandler(&ocgrpc.ServerHandler{}))
	}

	listener, err = net.Listen("tcp", config.AppConfig.Service.Address)
	if err != nil {
		return fmt.Errorf("failed to listen: %v", err)
	}

	grpcServer = grpc.NewServer(serverOptions...)
	grpc.StatsHandler(&ocgrpc.ServerHandler{})
	pb.RegisterAuthServer(grpcServer, &handler.AuthServer{})

	return nil
}

func Start() {
	log.Printf("gRPC listening on: %s\n", config.AppConfig.Service.Address)
	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

/*
	Observability: a measure of how well internal states of a system can be inferred from knowledge of its external outputs.
	OpenCensus is a set of libraries for collecting app metrics and distributed traces.
	It collects metrics from the target app and transfers the data to the backend of your choice in real time.
*/
func initOpenCensus() {
	go func() {
		mux := http.NewServeMux()
		// Package zpages implements a collection of HTML pages that display RPC stats and trace data
		zpages.Handle(mux, "/"+config.AppConfig.OpenCensus.StatsPage) // http://localhost:8081/debug/rpcz, http://localhost:8081/debug/tracez
		log.Fatal(http.ListenAndServe(config.AppConfig.OpenCensus.Address, mux))
	}()

	// Package view contains support for collecting and exposing aggregates over stats
	// PrintExporter is a stats and trace exporter that logs the exported data to the console.
	view.RegisterExporter(&exporter.PrintExporter{})

	// Register begins collecting data for the given views. Once a view is registered, it reports data to the registered exporters.
	if err := view.Register(ocgrpc.DefaultServerViews...); err != nil {
		log.Fatal(err)
	}

	log.Println("OpenCensus Enabled.")
	log.Printf("RPC Stats: http://%s/%s/rpcz\n", config.AppConfig.OpenCensus.Address, config.AppConfig.OpenCensus.StatsPage)
	log.Printf("Trace Spans: http://%s/%s/tracez\n", config.AppConfig.OpenCensus.Address, config.AppConfig.OpenCensus.StatsPage)
}
