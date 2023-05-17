package main

import (
	"embed"
	"github.com/algolia/algoliasearch-client-go/v3/algolia/search"
	"github.com/joho/godotenv"
	"github.com/tylermmorton/torque"
	"github.com/tylermmorton/torque/www/docsite/domain/content"
	"github.com/tylermmorton/torque/www/docsite/routes/docs"
	"io/fs"
	"log"
	"net/http"
	"os"
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

	algoliaAppId, ok := os.LookupEnv("ALGOLIA_APP_ID")
	if !ok {
		log.Fatalf("ALGOLIA_APP_ID not set in environment")
	}

	algoliaApiKey, ok := os.LookupEnv("ALGOLIA_API_KEY")
	if !ok {
		log.Fatalf("ALGOLIA_API_KEY not set in environment")
	}

	algoliaSearch := search.NewClient(algoliaAppId, algoliaApiKey)

	contentSvc, err := content.New(embeddedContent, algoliaSearch)
	if err != nil {
		log.Fatalf("failed to create content service: %+v", err)
	}

	staticAssets, err := fs.Sub(staticAssets, ".build/static")
	if err != nil {
		log.Fatalf("failed to create static assets filesystem: %+v", err)
	}

	r := torque.NewRouter(
		torque.WithFileSystemServer("/s", staticAssets),

		torque.WithRedirect("/", "/docs/", http.StatusTemporaryRedirect),

		torque.WithRouteModule("/docs/{pageName}", &docs.RouteModule{
			ContentSvc: contentSvc,
		}),
	)

	err = http.ListenAndServe(":8080", r)
	if err != nil {
		log.Fatalf("failed to start server: %+v", err)
	}
}
