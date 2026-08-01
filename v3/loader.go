package torque

import (
	"fmt"
	"net/http"
	"reflect"
	"sync"
)

func Load(req *http.Request, vm any) error {
	var vs visitorStack
	return executeLoadPlan(req, reflect.ValueOf(vm), &vs)
}

// visitorStack is a fixed length stack used to optimize the recursive visitor pattern.
// the maximum recursive depth is 32.
type visitorStack struct {
	buf [32]reflect.Type
	n   int
}

func (v *visitorStack) push(t reflect.Type) {
	if v.n >= len(v.buf) {
		panic("torque: loader depth exceeds 32")
	}
	v.buf[v.n] = t
	v.n++
}

func (v *visitorStack) pop() { v.n-- }

func (v *visitorStack) has(t reflect.Type) bool {
	for i := range v.n {
		if v.buf[i] == t {
			return true
		}
	}
	return false
}

// loadStep describes one direct field that is itself a Loader.
type loadStep struct {
	index     int  // field index within the struct
	isPointer bool // pointer field -> may need allocation before Load
}

// loadPlan is the precomputed, cached recipe for a single struct type. It only
// ever records type-level facts (indices), never instance values, which is why
// one plan safely serves every instance of a type.
type loadPlan struct {
	steps []loadStep
}

var (
	planCache  sync.Map // map[reflect.Type]*loadPlan
	loaderType = reflect.TypeOf((*Loader)(nil)).Elem()
)

// compileLoadPlan builds the plan for t once, then serves it from cache.
func compileLoadPlan(t reflect.Type) *loadPlan {
	if cached, ok := planCache.Load(t); ok {
		return cached.(*loadPlan)
	}

	plan := &loadPlan{}
	if t.Kind() == reflect.Struct {
		for i := 0; i < t.NumField(); i++ {
			f := t.Field(i)
			if !f.IsExported() {
				continue // unexported fields aren't reachable via Interface()
			}
			switch {
			case f.Type.Kind() == reflect.Ptr && f.Type.Implements(loaderType):
				plan.steps = append(plan.steps, loadStep{index: i, isPointer: true})
			case reflect.PointerTo(f.Type).Implements(loaderType):
				plan.steps = append(plan.steps, loadStep{index: i, isPointer: false})
			}
		}
	}

	actual, _ := planCache.LoadOrStore(t, plan)
	return actual.(*loadPlan)
}

// executeLoadPlan does a depth-first traversal of all Loader implementation on or within the given
// reflected value. Each Loader will populate the reflected value.
func executeLoadPlan(req *http.Request, v reflect.Value, visitor *visitorStack) error {
	elem := v.Elem()

	if elem.Kind() == reflect.Struct {
		st := elem.Type()
		visitor.push(st)

		for _, step := range compileLoadPlan(st).steps {
			fv := elem.Field(step.index)

			// The struct type this branch would descend into; guard cycles
			// before allocating so cyclic pointers are simply left nil.
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

			if err := executeLoadPlan(req, child, visitor); err != nil {
				return err
			}
		}

		visitor.pop()
	}

	// Depth first: children are hydrated above; now this node loads itself.
	if loader, ok := v.Interface().(Loader); ok {
		if err := loader.Load(req); err != nil {
			return fmt.Errorf("loading %s: %w", elem.Type().Name(), err)
		}
	}

	return nil
}
