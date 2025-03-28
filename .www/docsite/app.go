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

//go:generate rm -rf .bleve
//go:generate bash -c "(git rev-parse --short HEAD) > .gitrevision"

//go:embed .gitrevision
var revision string

//go:embed .build/static/*
var staticAssets embed.FS

//go:embed content/**
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

	contentFs, err := fs.Sub(embeddedContent, "content")
	if err != nil {
		log.Fatalf("failed to create content filesystem: %+v", err)
	}

	contentSvc, err := content.New(contentFs)
	if err != nil {
		log.Fatalf("failed to create content service: %+v", err)
	}

	r, err := torque.New[routes.ViewModel](&routes.Controller{
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
