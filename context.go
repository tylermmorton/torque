package torque

import (
	"context"
	"fmt"
	"net/http"
	"reflect"
	"sync"

	"github.com/gorilla/schema"
)

type contextKey string

const (
	errorKey         contextKey = "error"
	decoderKey       contextKey = "decoder"
	paramsContextKey contextKey = "params"

	routerMatchedContextKey contextKey = "router_matched"
	rootRouterKey           contextKey = "root_router"
	childContentKey         contextKey = "child_content"
	noOutletWrapKey         contextKey = "no_outlet_wrap"
)

func Provide[T any](req *http.Request, key any, value T) *http.Request {
	return req.WithContext(context.WithValue(req.Context(), key, value))
}

func Inject[T any](req *http.Request, key any) (T, error) {
	var noop T
	raw := req.Context().Value(key)
	if raw == nil {
		return noop, fmt.Errorf("inject: no value stored for key %v", key)
	}
	if value, ok := raw.(T); ok {
		return value, nil
	}
	return noop, fmt.Errorf("inject: expected %T but got %T", noop, raw)
}

func withError(req *http.Request, err error) *http.Request {
	return Provide(req, errorKey, err)
}

func UseError(req *http.Request) error {
	err, injectErr := Inject[error](req, errorKey)
	if injectErr != nil {
		return nil
	}
	return err
}

// contextStep describes one direct field that is itself a ContextProvider.
type contextStep struct {
	index     int
	isPointer bool
}

// contextPlan is the precomputed, cached recipe for a single struct type. It only
// ever records type-level facts (indices), never instance values.
type contextPlan struct {
	steps []contextStep
}

var (
	contextPlanCache    sync.Map
	contextProviderType = reflect.TypeOf((*ContextProvider)(nil)).Elem()
)

// compileContextPlan builds the plan for t once, then serves it from cache.
func compileContextPlan(t reflect.Type) *contextPlan {
	if cached, ok := contextPlanCache.Load(t); ok {
		return cached.(*contextPlan)
	}

	plan := &contextPlan{}
	if t.Kind() == reflect.Struct {
		for i := 0; i < t.NumField(); i++ {
			f := t.Field(i)
			if !f.IsExported() {
				continue
			}
			switch {
			case f.Type.Kind() == reflect.Ptr && f.Type.Implements(contextProviderType):
				plan.steps = append(plan.steps, contextStep{index: i, isPointer: true})
			case reflect.PointerTo(f.Type).Implements(contextProviderType):
				plan.steps = append(plan.steps, contextStep{index: i, isPointer: false})
			}
		}
	}

	actual, _ := contextPlanCache.LoadOrStore(t, plan)
	return actual.(*contextPlan)
}

func Context(req *http.Request, vm any) *http.Request {
	var vs visitorStack
	return executeContextPlan(req, reflect.ValueOf(vm), &vs)
}

func executeContextPlan(req *http.Request, v reflect.Value, visitor *visitorStack) *http.Request {
	elem := v.Elem()

	if elem.Kind() == reflect.Struct {
		st := elem.Type()
		visitor.push(st)

		// Pre-order: call Context on current node before descending into children.
		if cp, ok := v.Interface().(ContextProvider); ok {
			req = cp.Context(req)
		}

		for _, step := range compileContextPlan(st).steps {
			fv := elem.Field(step.index)

			childStruct := fv.Type()
			if childStruct.Kind() == reflect.Ptr {
				childStruct = childStruct.Elem()
			}
			if visitor.has(childStruct) {
				continue
			}

			var child reflect.Value
			if step.isPointer {
				if fv.IsNil() {
					fv.Set(reflect.New(fv.Type().Elem()))
				}
				child = fv
			} else {
				child = fv.Addr()
			}

			// Each sibling receives the parent's req; one child's result
			// does not carry over to the next sibling.
			executeContextPlan(req, child, visitor)
		}

		visitor.pop()
	}

	return req
}

func InjectDecoder(req *http.Request) (*schema.Decoder, bool) {
	d, err := Inject[*schema.Decoder](req, decoderKey)
	return d, err == nil
}
