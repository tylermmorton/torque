package torque

import (
	"context"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"path/filepath"
	"strings"
)

const (
	parameterKey = "{}"
)

type Router interface {
	http.Handler

	Handle(pattern string, handler http.Handler)
	HandleFileSystem(pattern string, fs fs.FS)
	HandleRedirect(pattern string, location string, status int)

	Use(mw Middleware)
	ProvideContext(key any, value any)
	ProvideTemplate(name string, tp TemplateProvider) error

	Match(method, pattern string) (http.Handler, PathParams, bool)
}

type Middleware func(http.Handler) http.Handler

type RouterOption func(*routerImpl)

// DisableRootLayout disables the default PageLayout root layout.
func DisableRootLayout() RouterOption {
	return func(r *routerImpl) {
		r.rootLayout = nil
	}
}

type trieNode struct {
	segment    string
	parent     *trieNode
	children   map[string]*trieNode
	handlers   map[string]http.Handler
	isParam    bool
	paramName  string
	contextMap map[any]any
}

type routerImpl struct {
	h          Handler
	rootLayout Handler
	contextMap map[any]any
	templates  map[string]TemplateProvider

	root   *trieNode
	prefix string
}

func NewRouter(opts ...RouterOption) Router {
	r := &routerImpl{
		rootLayout: MustNewHandler[PageLayout](),
		contextMap: make(map[any]any),
		templates:  make(map[string]TemplateProvider),
		root: &trieNode{
			children: make(map[string]*trieNode),
			handlers: map[string]http.Handler{},
		},
	}
	for _, opt := range opts {
		opt(r)
	}
	return r
}

func (r *routerImpl) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	// The outermost router injects itself so outlet funcs can dispatch sub-requests.
	if ctx.Value(rootRouterKey) == nil {
		ctx = context.WithValue(ctx, rootRouterKey, Router(r))
		req = req.WithContext(ctx)
	}

	h, params, ok := r.Match(req.Method, req.URL.Path)
	if !ok {
		// Fall back to the owning handler (serves the handler's own URL when used standalone).
		if r.h != nil {
			ctx = context.WithValue(req.Context(), paramsContextKey, PathParams{})
			ctx = context.WithValue(ctx, routerMatchedContextKey, true)
			r.h.ServeHTTP(w, req.WithContext(ctx))
			return
		}
		http.NotFound(w, req)
		return
	}

	ctx = context.WithValue(req.Context(), paramsContextKey, params)
	ctx = context.WithValue(ctx, routerMatchedContextKey, true)
	h.ServeHTTP(w, req.WithContext(ctx))
}

func (r *routerImpl) Handle(path string, h http.Handler) {
	r.handleMethod("*", path, h)
}

// handleMethod registers a handler or merges a routerImpl if passed.
func (r *routerImpl) handleMethod(method, path string, h http.Handler) {
	var (
		fullPath = filepath.Join(r.prefix, path)
		segments = strings.Split(fullPath, "/")
		node     = r.root
	)
	for _, segment := range segments {
		if segment == "" {
			continue
		}

		isParam := strings.HasPrefix(segment, "{") && strings.HasSuffix(segment, "}")
		var key string
		if isParam {
			key = parameterKey
		} else {
			key = segment
		}

		if _, exists := node.children[key]; !exists {
			node.children[key] = &trieNode{
				segment:  segment,
				parent:   node,
				children: make(map[string]*trieNode),
				handlers: make(map[string]http.Handler),
				isParam:  isParam,
				paramName: func() string {
					if isParam {
						return segment[1 : len(segment)-1] // Extract param name (e.g., userId from {userId})
					}
					return ""
				}(),
			}
		}

		node = node.children[key]
	}

	// Store the handler at the final node for the given method (e.g., GET)
	node.handlers[method] = h

	if handler, ok := h.(Handler); ok {
		rootParent := r.rootLayout
		if rootParent == nil {
			rootParent = r.h
		}

		if rootParent != nil {
			var parentMostHandler = handler
			for {
				if parentMostHandler.GetParent() != nil {
					parentMostHandler = parentMostHandler.GetParent()
				} else {
					break
				}
			}
			parentMostHandler.setParent(rootParent)
		}

		for name, tp := range r.templates {
			if err := handler.provideTemplate(name, tp); err != nil {
				// TODO: How can we prevent this scenario? Surface this error earlier
				//  instead of when we're actually handling the request
				log.Printf("torque: failed to inject supplemental template %q into handler: %s", name, err)
			}
		}

		// "merge-up" the radix sub-trie from the child router. when this handler's internal
		// router is ever executed it will need to know about its children during Router.Match.
		if handler.getRouter() != nil {
			var childRouter = handler.getRouter()
			node.contextMap = childRouter.contextMap
			for key, child := range childRouter.root.children {
				node.children[key] = child
			}
			if len(childRouter.root.handlers) > 0 {
				node.handlers[method] = childRouter.root.handlers[method]
			}
		}
	}
}

// Match finds a handler based on the method and path
func (r *routerImpl) Match(method, path string) (http.Handler, PathParams, bool) {
	params := make(map[string]string)
	segments := strings.Split(path, "/")

	// Pre-seed with the root router's own context map.
	var contextMaps []map[any]any
	if len(r.contextMap) > 0 {
		contextMaps = append(contextMaps, r.contextMap)
	}

	// Traverse the radix trie to find the matching handler
	node := r.root
	for _, segment := range segments {
		if segment == "" {
			continue
		}

		if child, exists := node.children[segment]; exists {
			node = child
		} else if paramChild, exists := node.children["{}"]; exists {
			node = paramChild
			params[node.paramName] = segment
		} else if wildcardChild, exists := node.children["*"]; exists {
			node = wildcardChild
			break
		} else {
			return nil, nil, false
		}

		if len(node.contextMap) > 0 {
			contextMaps = append(contextMaps, node.contextMap)
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
		if len(contextMaps) > 0 {
			handler = wrapWithContext(handler, contextMaps)
		}
		return handler, params, true
	}

	return nil, nil, false
}

func wrapWithContext(h http.Handler, maps []map[any]any) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		ctx := req.Context()
		for _, m := range maps {
			for k, v := range m {
				ctx = context.WithValue(ctx, k, v)
			}
		}
		h.ServeHTTP(w, req.WithContext(ctx))
	})
}

func (r *routerImpl) HandleFileSystem(pattern string, fs fs.FS) {}

func (r *routerImpl) Use(mw Middleware) {}

func (r *routerImpl) ProvideContext(key any, value any) {
	r.contextMap[key] = value
}

func (r *routerImpl) ProvideTemplate(name string, tp TemplateProvider) error {
	if _, ok := r.templates[name]; ok {
		return fmt.Errorf("template with name %q already defined", name)
	}

	r.templates[name] = tp
	return nil
}

//func (r *router) HandleFileSystem(pattern string, fs fs.FS) {
//	pattern = strings.TrimSuffix(pattern, "/*")
//
//	if r.h.GetMode() == ModeDevelopment {
//		logFileSystem(fs)
//	}
//
//	r.handleMethod("GET", pattern+"/*", NoOutlet(http.StripPrefix(pattern, http.FileServer(http.FS(fs)))))
//}

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

type noWrapHandler func(http.ResponseWriter, *http.Request)

func (h noWrapHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	h(w, req)
}

// NoOutlet indicates to the Router that the given http.Handler should not be wrapped when
// adding it via Handle. This is useful when you want to pass a vanilla http.Handler to
// a Router that shouldn't be wrapped by its parent's output.
func NoOutlet(h http.Handler) http.Handler {
	return noWrapHandler(func(w http.ResponseWriter, req *http.Request) {
		h.ServeHTTP(w, req)
	})
}

func (r *routerImpl) HandleRedirect(pattern string, url string, status int) {
	r.Handle(pattern, NoOutlet(http.RedirectHandler(url, status)))
}
