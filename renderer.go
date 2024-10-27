package torque

import (
	"github.com/pkg/errors"
	"net/http"
)

var (
	// ErrRenderFnNotDefined is returned by VaryRender when the given header value is not
	// found in the given cases map and a default case is not provided.
	ErrRenderFnNotDefined = errors.New("render function not defined for header value")
)

// RenderFn is a function that renders a response to the given http.ResponseWriter.
type RenderFn[T ViewModel] func(wr http.ResponseWriter, req *http.Request, vm T) error

//
//// VaryDefault is a special key that can be used in the cases map of VaryRender to
//// indicate that the given RenderFn should be used as the default case.
//const VaryDefault key = "default"
//
//// VaryRender is a helper function for rendering different responses based on the given
//// header key. The header key is used to look up a RenderFn in the given cases map.
////
//// Sometimes it is useful to have a default case that is used when the header value is not
//// found in the map. To do this, add VaryDefault as the key in the case map.
////
//// If the header value is not found in the map and a default is not provided, then VaryRender
//// returns ErrRenderFnNotDefined.
////
//// The given header is also written to the Vary header of the response to indicate to the
//// browser cache that responses from this endpoint may vary.
//func VaryRender(wr http.ResponseWriter, req *http.Request, header string, cases map[any]RenderFn) error {
//	// Tell the cache that the response can vary based on the request header
//	wr.Header().Set("Vary", header)
//
//	value := req.Header.Get(header)
//	if fn, ok := cases[value]; ok {
//		return fn(wr, req)
//	} else if fn, ok := cases[VaryDefault]; ok {
//		return fn(wr, req)
//	} else {
//		return errors.WithMessage(ErrRenderFnNotDefined, fmt.Sprintf("render function not found for header value %q", value))
//	}
//}
