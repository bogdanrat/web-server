package main

import (
	"github.com/bogdanrat/web-server/app"
	"log"
)

func main() {
	if err := app.Init(); err != nil {
		log.Fatal(err)
	}

	app.Start()
}
