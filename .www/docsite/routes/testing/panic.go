package testing

import (
	"errors"
	"net/http"
)

//for testing purpose only
type RouteModule struct {
}

func (rm *RouteModule) Load(req *http.Request) (interface{}, error) {
	panic(errors.New("Oh no!"))
}
