package main

import (
	"log"
	"net/http"

	"github.com/tylermmorton/torque/.www/docsite"
)

func main() {
	app, err := docsite.NewDocSite()
	if err != nil {
		log.Fatalf("failed to create app: %v", err)
	}

	err = http.ListenAndServe(":8080", app)
	if err != nil {
		log.Fatalf("failed start http server: %v", err)
	}
}
