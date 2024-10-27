package torque

import (
	"context"
	"html/template"
	"io/fs"
	"net/http"
	"net/http/httptest"
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
	parent    *trieNode
	children  map[string]*trieNode
	handlers  map[string]http.Handler
	isParam   bool
	paramName string
	isOutlet  bool
}

type router struct {
	root   *trieNode
	prefix string
}

func createRouter[T ViewModel](h *handlerImpl[T], ctl Controller, isOutlet bool) *router {
	r := &router{
		prefix: h.path,
		root: &trieNode{
			children: make(map[string]*trieNode),
			handlers: map[string]http.Handler{
				"*": h,
			},
			isOutlet: isOutlet,
		},
	}

	if rp, ok := ctl.(RouterProvider); ok {
		rp.Router(r)
	}

	return r
}

func (r *router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	handler, params, ok := r.Match(req.Method, req.URL.Path)
	if !ok {
		http.NotFound(w, req)
		return
	}

	// Add route parameters to the request context
	ctx := req.Context()
	for key, value := range params {
		ctx = context.WithValue(ctx, key, value)
	}

	handler.ServeHTTP(w, req.WithContext(ctx))
}

func (r *router) Handle(path string, h http.Handler) {
	r.handleMethod("*", path, h)
}

// handleMethod registers a handler or merges a router if passed.
func (r *router) handleMethod(method, path string, h http.Handler) {
	fullPath := filepath.Join(r.prefix + path)

	// If another router is passed, create a parent/child relationship between them
	if handler, ok := h.(Handler); ok && handler.getRouter() != nil {
		child := handler.getRouter().root
		segment := strings.TrimPrefix(path, "/")

		r.root.children[segment] = child
		child.parent = r.root

		return
	}

	handler, ok := h.(http.Handler)
	if !ok {
		panic("invalid handler or router passed to Handle")
	}

	// Split the full path into segments
	segments := strings.Split(fullPath, "/")

	// Special case for root path "/"
	if fullPath == rootKey || (len(segments) == 2 && segments[1] == "") {
		r.root.handlers = make(map[string]http.Handler)
		r.root.handlers[method] = handler
		return
	}

	// Traverse through the segments
	node := r.root
	for _, seg := range segments {
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
		} else {
			return nil, nil, false
		}
	}

	// Return the handler if it exists for the given method or wildcard.
	// Before returning, recursively wrap it with any parent outlets
	if handler, ok := node.handlers[method]; ok {
		return wrapWithParentOutlet(handler, node, method), params, true
	} else if handler, ok := node.handlers["*"]; ok {
		return wrapWithParentOutlet(handler, node, method), params, true
	}
	return nil, nil, false
}

func wrapWithParentOutlet(childHandler http.Handler, node *trieNode, method string) http.Handler {
	if node.parent != nil && node.parent.isOutlet {
		var parentHandler http.Handler
		if h, exists := node.parent.handlers[method]; exists {
			parentHandler = h
		} else if h, exists = node.parent.handlers["*"]; exists {
			parentHandler = h
		} else {
			panic("this should not happen")
		}
		return wrapWithParentOutlet(http.HandlerFunc(func(wr http.ResponseWriter, req *http.Request) {
			// Indicate to any handlers they should not attempt to route the request
			// using their internal router, and instead just serve the request
			req = req.WithContext(context.WithValue(req.Context(), outletKey, true))

			var (
				childReq   = req
				childResp  = httptest.NewRecorder()
				parentReq  = req.Clone(req.Context())
				parentResp = httptest.NewRecorder()
			)

			// child before parent, because it can set additional context
			// while handling the request
			childHandler.ServeHTTP(childResp, childReq)
			parentHandler.ServeHTTP(parentResp, parentReq.WithContext(childReq.Context()))

			t := template.Must(template.New("outlet").Parse(parentResp.Body.String()))

			err := t.Execute(wr, template.HTML(childResp.Body.String()))
			if err != nil {
				panic(err)
			}
		}), node.parent, method)
	} else {
		return childHandler
	}
}

//func (r *router) Use(middleware ...Middleware) {
//	r.middleware = append(r.middleware, middleware...)
//}

// // RouteParam returns the named route parameter from the request url
//
//	func RouteParam(req *http.Request, name string) string {
//		return chi.URLParam(req, name)
//	}
//
//	type Router interface {
//		chi.Router
//
//		HandleFileSystem(pattern string, fs fs.FS)
//	}
//
//	type routerImpl struct {
//		chi.Router
//		Handler Handler
//	}
//
//	func logRoutes(prefix string, r []chi.Route) {
//		for _, route := range r {
//			pattern := filepath.Join(prefix, route.Pattern)
//			log.Printf("Route: %s\n", pattern)
//			if route.SubRoutes != nil {
//				logRoutes(pattern, route.SubRoutes.Routes())
//			}
//		}
//	}
//

func (r *router) HandleFileSystem(pattern string, fs fs.FS) {
	//r.Router.Route(pattern, func(r chi.Router) {
	//	r.Get("/*", func(wr http.ResponseWriter, req *http.Request) {
	//		log.Printf("[FileSystem] %s", req.URL.Path)
	//		http.StripPrefix(pattern, http.FileServer(http.FS(fs))).ServeHTTP(wr, req)
	//	})
	//})
	//if r.Handler.GetMode() == ModeDevelopment {
	//	log.Printf("-- HandleFileSystem(%s) --", pattern)
	//	logFileSystem(fs)
	//}
}

//
//func logFileSystem(fsys fs.FS) {
//	var walkFn func(path string, d fs.DirEntry, err error) error
//
//	walkFn = func(path string, d fs.DirEntry, err error) error {
//		if err != nil {
//			return err
//		} else if d.IsDir() {
//			log.Printf("Dir: %s", path)
//		} else {
//			log.Printf("File: %s", path)
//		}
//		return nil
//	}
//
//	err := fs.WalkDir(fsys, ".", walkFn)
//	if err != nil {
//		panic(err)
//	}
//}
