package main

import (
	"embed"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"

	algolia "github.com/algolia/algoliasearch-client-go/v3/algolia/search"
	"github.com/joho/godotenv"
	"github.com/tylermmorton/torque"
	"github.com/tylermmorton/torque/.www/docsite/routes/docs"
	"github.com/tylermmorton/torque/.www/docsite/services/content"
)

//go:generate tmpl bind ./... --outfile=tmpl.gen.go

//go:embed .build/static/*
var staticAssets embed.FS

//go:embed content/docs/*
var embeddedContent embed.FS

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Printf("failed to load env: %+v", err)
	}

	algoliaAppId, ok := os.LookupEnv("ALGOLIA_APP_ID")
	if !ok {
		log.Fatalf("ALGOLIA_APP_ID not set in environment")
	}

	algoliaApiKey, ok := os.LookupEnv("ALGOLIA_API_KEY")
	if !ok {
		log.Fatalf("ALGOLIA_API_KEY not set in environment")
	}

	algoliaClient := algolia.NewClient(algoliaAppId, algoliaApiKey)

	contentSvc, err := content.New(embeddedContent, algoliaClient)
	if err != nil {
		log.Fatalf("failed to create content service: %+v", err)
	}

	var assetHandler torque.RouteComponent
	if os.Getenv("EMBED_ASSETS") == "true" {
		staticAssets, err := fs.Sub(staticAssets, ".build/static")
		if err != nil {
			log.Fatalf("failed to create static assets filesystem: %+v", err)
		}
		assetHandler = torque.WithFileSystemServer("/s", staticAssets)
	} else {
		assetHandler = torque.WithFileServer("/s", ".build/static")
	}

	r := torque.NewRouter(
		assetHandler,

		torque.WithRouteModule("/{pageName}", &docs.RouteModule{ContentSvc: contentSvc}),
		torque.WithRouteModule("/panic", &testing.RouteModule{}),
		torque.WithRedirect("/", "/getting-started", http.StatusTemporaryRedirect),
	)

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
