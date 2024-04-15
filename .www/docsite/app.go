package main

import (
	"embed"
	"fmt"
	"github.com/tylermmorton/torque/.www/docsite/routes"
	"io/fs"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"github.com/tylermmorton/torque"
	"github.com/tylermmorton/torque/.www/docsite/services/content"
)

//go:embed .build/static/*
var staticAssets embed.FS

//go:embed content/docs/*
var embeddedContent embed.FS

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Printf("failed to load env: %+v", err)
	}

	assetsFs, err := fs.Sub(staticAssets, ".build/static")
	if err != nil {
		log.Fatalf("failed to create static assets filesystem: %+v", err)
	}

	contentSvc, err := content.New(embeddedContent, nil)
	if err != nil {
		log.Fatalf("failed to create content service: %+v", err)
	}

	r, err := torque.New[routes.IndexView](&routes.IndexHandlerModule{
		StaticAssets:   assetsFs,
		ContentService: contentSvc,
	})
	if err != nil {
		log.Fatalf("failed to create controller: %+v", err)
	}

	var host, port = os.Getenv("HOST_ADDR"), os.Getenv("HOST_PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Starting server on %s:%s", host, port)
	err = http.ListenAndServe(fmt.Sprintf("%s:%s", host, port), r)
	if err != nil {
		log.Fatalf("failed to start server: %+v", err)
	}
}
