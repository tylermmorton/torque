package main

import (
	"github.com/tylermmorton/torque"
	"github.com/tylermmorton/torque/www/docsite/domain/content"
	"github.com/tylermmorton/torque/www/docsite/endpoints/docs"
	"log"
	"net/http"
)

func main() {
	contentSvc, err := content.New()
	if err != nil {
		log.Fatalf("failed to create content service: %+v", err)
	}

	app := torque.NewApp(
		torque.WithRedirect("/", "/docs/index", http.StatusTemporaryRedirect),

		// TODO(tylermorton): Refactor this to be more ergonomic. This will break in build mode
		// because the static files will be in a different directory.
		// Perhaps experiment with embedded file systems.
		torque.WithFileServer("/s", "./.build/static"),

		torque.WithHttp("/docs/{pageName}", &docs.RouteModule{
			ContentSvc: contentSvc,
		}),
	)

	err = http.ListenAndServe(":8080", app)
	if err != nil {
		log.Fatalf("failed to start server: %+v", err)
	}
}
