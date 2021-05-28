package main

import (
	"github.com/bogdanrat/web-server/service/storage/app"
	_ "google.golang.org/grpc/encoding/gzip"
	"log"
)

func main() {
	if err := app.Init(); err != nil {
		log.Fatal(err)
	}

	app.Start()
}
