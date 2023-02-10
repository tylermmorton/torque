package main

import (
	"github.com/tylermmorton/torque"
	. "github.com/tylermmorton/torque/examples/home"
	. "github.com/tylermmorton/torque/examples/login"

	"log"
	"net/http"
)

func main() {
	app := torque.NewApp(
		torque.Route("/",
			&HomeRoute{},
			torque.WithGuard(nil),
		),
		torque.Route("/login",
			&LoginRoute{},
			torque.WithGuard(nil),
		),
	)

	err := http.ListenAndServe(":8080", app)
	if err != nil {
		log.Fatalf("failed to run app server: %v", err)
	}
}
