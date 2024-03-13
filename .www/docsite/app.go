package main

import (
	"embed"
	"fmt"
	"github.com/tylermmorton/torque/pkg/plugins/templ"
	"io/fs"
	"log"
	"net/http"
	"os"

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

// docsApp represents the root of the doc site application.
type docsApp struct {
	StaticAssets   fs.FS
	ContentService content.Service
}

// Render is the handler for the root of the site ... just redirect to getting started
func (*docsApp) Render(wr http.ResponseWriter, req *http.Request, loaderData any) error {
	http.Redirect(wr, req, "/getting-started", http.StatusFound)
	return nil
}

func (d *docsApp) Router(r torque.Router) {
	r.HandleModule("/{pageName}", &docs.RouteModule{ContentService: d.ContentService})
	r.HandleFileSystem("/s", d.StaticAssets)
}

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

	r := torque.NewViewController(&docsApp{
		StaticAssets:   assetsFs,
		ContentService: contentSvc,
	}, torque.WithPlugin(&templ.Plugin{}))

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
