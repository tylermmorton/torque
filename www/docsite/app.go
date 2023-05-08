package main

import (
	"github.com/tylermmorton/torque"
	"github.com/tylermmorton/torque/www/docsite/domain/content"
	"github.com/tylermmorton/torque/www/docsite/endpoints/index"
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
		torque.WithFileServer("/s", "./.build/static"),
		torque.WithHttp("/docs/{pageName}", &index.RouteModule{
			ContentSvc: contentSvc,
		}),
	)

	err = http.ListenAndServe(":8080", app)
	if err != nil {
		log.Fatalf("failed to start server: %+v", err)
	}
}
