package torque

import (
	"fmt"
	"github.com/pkg/errors"
	"net/http"
)

// TODO(v1.0) I like SplitRender but does SwitchRender or VaryRender make more sense semantically?
// TODO(idea) https://github.com/carlmjohnson/truthy <- Add 'truthy' testing to SplitRender to allow for more flexible header values

var (
	// ErrRenderFnNotDefined is returned by SplitRender when the given header value is not
	// found in the given cases map and a default case is not provided.
	ErrRenderFnNotDefined = errors.New("split render function not defined for header value")
)

// RenderFn is a function that renders a response to the given http.ResponseWriter.
type RenderFn = func(wr http.ResponseWriter, req *http.Request) error

// SplitRenderDefault is a special key that can be used in the cases map of SplitRender to
// indicate that the given RenderFn should be used as the default case.
const SplitRenderDefault key = "default"

// SplitRender is a helper function for rendering different responses based on the given
// header key. The header key is used to look up a RenderFn in the given cases map.
//
// Sometimes it is useful to have a default case that is used when the header value is not
// found in the map. To do this, add SplitRenderDefault as the key in the case map.
//
// If the header value is not found in the map and a default is not provided, then SplitRender
// returns ErrRenderFnNotDefined.
//
// The given header is also written to the Vary header of the response to indicate to the
// browser cache that responses from this endpoint may vary.
func SplitRender(wr http.ResponseWriter, req *http.Request, header string, cases map[any]RenderFn) error {
	// Tell the cache that the response can vary based on the request header
	wr.Header().Set("Vary", header)

	value := req.Header.Get(header)
	if fn, ok := cases[value]; ok {
		return fn(wr, req)
	} else if fn, ok := cases[SplitRenderDefault]; ok {
		return fn(wr, req)
	} else {
		return errors.WithMessage(ErrRenderFnNotDefined, fmt.Sprintf("render function not found for header value %q", value))
	}
}
