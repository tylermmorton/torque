package main

import (
	"net/http"

	"github.com/tylermmorton/torque"
	"github.com/tylermmorton/torque/examples/quick-start/homepage"
)

func main() {
	h := torque.MustNew[homepage.ViewModel](&homepage.Controller{})
	http.ListenAndServe(":9001", h)
}
