package main

import (
	"github.com/bogdanrat/web-server/service/core/app"
	"log"
)

func main() {
	if err := app.Init(); err != nil {
		log.Fatal(err)
	}

	app.Start()
}
