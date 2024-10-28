package torque

import (
	"context"
	"io/fs"
	"log"
	"net/http"
	"path/filepath"
	"strings"
)

const (
	rootKey      = "/"
	parameterKey = "{}"
)

type Router interface {
	http.Handler

	Handle(pattern string, handler http.Handler)
	HandleFileSystem(pattern string, fs fs.FS)

	Match(method, path string) (http.Handler, map[string]string, bool)
}

type Middleware func(http.Handler) http.Handler

type trieNode struct {
	segment   string
	parent    *trieNode
	children  map[string]*trieNode
	handlers  map[string]http.Handler
	isParam   bool
	paramName string
}

type router struct {
	h      Handler
	root   *trieNode
	prefix string
}

func createRouter[T ViewModel](h *handlerImpl[T], ctl Controller) *router {
	r := &router{
		h:      h,
		prefix: h.path,
		root: &trieNode{
			children: make(map[string]*trieNode),
			handlers: map[string]http.Handler{"*": h},
		},
	}

	if rp, ok := ctl.(RouterProvider); ok {
		rp.Router(r)
	}

	return r
}

func (r *router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	h, params, ok := r.Match(req.Method, req.URL.Path)
	if !ok {
		http.NotFound(w, req)
		return
	}

	ctx := req.Context()
	ctx = context.WithValue(ctx, paramsContextKey, params)

	h.ServeHTTP(w, req.WithContext(ctx))
}

func (r *router) Handle(path string, h http.Handler) {
	r.handleMethod("*", path, h)
}

// handleMethod registers a handler or merges a router if passed.
func (r *router) handleMethod(method, path string, h http.Handler) {
	fullPath := filepath.Join(r.prefix, path)
	segments := strings.Split(fullPath, "/")

	handler, ok := h.(http.Handler)
	if !ok {
		panic("invalid handler or router passed to Handle")
	}

	var seg string
	node := r.root
	for _, seg = range segments {
		if seg == "" {
			continue
		}

		isParam := strings.HasPrefix(seg, "{") && strings.HasSuffix(seg, "}")
		var key string
		if isParam {
			key = parameterKey
		} else {
			key = seg
		}

		if _, exists := node.children[key]; !exists {
			node.children[key] = &trieNode{
				segment:  seg,
				parent:   node,
				children: make(map[string]*trieNode),
				handlers: make(map[string]http.Handler),
				isParam:  isParam,
				paramName: func() string {
					if isParam {
						return seg[1 : len(seg)-1] // Extract param name (e.g., userId from {userId})
					}
					return ""
				}(),
			}
		}

		node = node.children[key]
	}

	// Store the handler at the final node for the given method (e.g., GET)
	node.handlers[method] = handler

	// If another Handler was passed, create a relationship between the parent and child
	if h, ok := h.(Handler); ok {
		if r.h.HasOutlet() {
			h.setParent(r.h)
		}

		// "merge-up" the radix sub-trie from the child router. when this handler's internal
		// router is ever executed it will need to know about its children during Router.Match.
		if h.getRouter() != nil {
			var childRouter = h.getRouter().root
			for key, child := range childRouter.children {
				node.children[key] = child
			}
		}
	}
}

// Match finds a handler based on the method and path
func (r *router) Match(method, path string) (http.Handler, map[string]string, bool) {
	params := make(map[string]string)
	segments := strings.Split(path, "/")

	// Traverse the radix trie to find the matching handler
	node := r.root
	for _, seg := range segments {
		if seg == "" {
			continue
		}

		if child, exists := node.children[seg]; exists {
			node = child
		} else if paramChild, exists := node.children["{}"]; exists {
			node = paramChild
			params[node.paramName] = seg
		} else if wildcardChild, exists := node.children["*"]; exists {
			node = wildcardChild
			break
		} else {
			return nil, nil, false
		}
	}

	// Return the handler if it exists for the given method or wildcard.
	var handler http.Handler
	if h, ok := node.handlers[method]; ok {
		handler = h
	} else if h, ok := node.handlers["*"]; ok {
		handler = h
	}

	if handler != nil {
		return handler, params, true
	} else {
		return nil, nil, false
	}
}

func (r *router) HandleFileSystem(pattern string, fs fs.FS) {
	pattern = strings.TrimSuffix(pattern, "/*")

	if r.h.GetMode() == ModeDevelopment {
		logFileSystem(fs)
	}

	r.handleMethod("GET", pattern+"/*", http.StripPrefix(pattern, http.FileServer(http.FS(fs))))
}

func logFileSystem(fsys fs.FS) {
	var walkFn func(path string, d fs.DirEntry, err error) error

	walkFn = func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		} else if d.IsDir() {
			log.Printf("Dir: %s", path)
		} else {
			log.Printf("File: %s", path)
		}
		return nil
	}

	err := fs.WalkDir(fsys, ".", walkFn)
	if err != nil {
		panic(err)
	}
}

func RouteParam(req *http.Request, key string) string {
	if params, ok := req.Context().Value(paramsContextKey).(map[string]string); ok {
		if val, exists := params[key]; exists {
			return val
		}
	}
	return ""
}
