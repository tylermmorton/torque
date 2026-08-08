package torque

import (
	htmltemplate "html/template"
	"io"
	"net/http"
	"reflect"
	"strings"
	"sync"
	"text/template"

	"github.com/tylermmorton/torque/pkg/templates/html"
)

type StyleSheet[T StyleSheetProvider] interface {
	Render(wr io.Writer, data any) error
}

type stylesheetImpl[T StyleSheetProvider] struct {
	template *template.Template
}

// CompileStyleSheet compiles the CSS template returned by ssp.StyleSheet() once at creation time.
// TODO: unify FuncMapProvider across all template-based systems (StyleSheet, Template, etc.)
func CompileStyleSheet[T StyleSheetProvider](ssp T) (StyleSheet[T], error) {
	t, err := template.New("stylesheet").Parse(ssp.StyleSheet())
	if err != nil {
		return nil, err
	}
	return &stylesheetImpl[T]{template: t}, nil
}

func (s *stylesheetImpl[T]) Render(wr io.Writer, data any) error {
	return s.template.Execute(wr, data)
}

// stylesheetStep records the index of a struct field to visit during stylesheet execution.
type stylesheetStep struct {
	index     int
	isPointer bool
}

// stylesheetPlan is the precomputed, per-type recipe for stylesheet execution. It stores
// a compiled root stylesheet (if the type itself implements StyleSheetProvider) and field
// indices to recurse into. One plan safely serves every instance of a type.
type stylesheetPlan struct {
	root  StyleSheet[StyleSheetProvider]
	steps []stylesheetStep
}

var (
	stylesheetPlanCache    sync.Map
	stylesheetProviderType = reflect.TypeOf((*StyleSheetProvider)(nil)).Elem()
)

func compileStylesheetPlan(t reflect.Type) (*stylesheetPlan, error) {
	if cached, ok := stylesheetPlanCache.Load(t); ok {
		return cached.(*stylesheetPlan), nil
	}

	plan := &stylesheetPlan{}

	if reflect.PointerTo(t).Implements(stylesheetProviderType) {
		zeroVal := reflect.New(t).Interface().(StyleSheetProvider)
		ss, err := CompileStyleSheet(zeroVal)
		if err != nil {
			return nil, err
		}
		plan.root = ss
	}

	if t.Kind() == reflect.Struct {
		for i := 0; i < t.NumField(); i++ {
			f := t.Field(i)
			if !f.IsExported() {
				continue
			}
			switch {
			case f.Type.Kind() == reflect.Ptr && f.Type.Implements(stylesheetProviderType):
				plan.steps = append(plan.steps, stylesheetStep{index: i, isPointer: true})
			case reflect.PointerTo(f.Type).Implements(stylesheetProviderType):
				plan.steps = append(plan.steps, stylesheetStep{index: i, isPointer: false})
			}
		}
	}

	actual, _ := stylesheetPlanCache.LoadOrStore(t, plan)
	return actual.(*stylesheetPlan), nil
}

func executeStylesheetPlan(req *http.Request, v reflect.Value, vs *visitorStack) (*http.Request, error) {
	elem := v
	if elem.Kind() == reflect.Ptr {
		elem = elem.Elem()
	}
	if elem.Kind() != reflect.Struct {
		return req, nil
	}

	st := elem.Type()
	vs.push(st)

	plan, err := compileStylesheetPlan(st)
	if err != nil {
		return req, err
	}

	// Pre-order: execute root stylesheet before descending into fields.
	if plan.root != nil {
		var buf strings.Builder
		if err := plan.root.Render(&buf, v.Interface()); err != nil {
			return req, err
		}
		if css := buf.String(); css != "" {
			req = ProvideInlineStyles(req, html.StyleTag{Content: htmltemplate.CSS(css)})
		}
	}

	for _, step := range plan.steps {
		fv := elem.Field(step.index)

		childType := fv.Type()
		if childType.Kind() == reflect.Ptr {
			childType = childType.Elem()
		}
		if vs.has(childType) {
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

		req, err = executeStylesheetPlan(req, child, vs)
		if err != nil {
			return req, err
		}
	}

	vs.pop()
	return req, nil
}

// ApplyStyleSheets traverses vm collecting CSS from all StyleSheetProvider implementations
// found on the value and its fields, injecting each into the request context.
func ApplyStyleSheets(req *http.Request, vm any) (*http.Request, error) {
	var vs visitorStack
	return executeStylesheetPlan(req, reflect.ValueOf(vm), &vs)
}
