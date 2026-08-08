package torque_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tylermmorton/torque"
)

type selfLoader struct {
	Value string
}

func (s *selfLoader) Load(req *http.Request) error {
	s.Value = "loaded"
	return nil
}

type orderLeaf struct {
	order *[]string
}

func (l *orderLeaf) Load(req *http.Request) error {
	*l.order = append(*l.order, "orderLeaf")
	return nil
}

type orderMid struct {
	order *[]string
	Leaf  orderLeaf
}

func (m *orderMid) Load(req *http.Request) error {
	*m.order = append(*m.order, "orderMid")
	return nil
}

type orderRoot struct {
	order *[]string
	Mid   orderMid
}

func (r *orderRoot) Load(req *http.Request) error {
	*r.order = append(*r.order, "orderRoot")
	return nil
}

type ptrChild struct {
	Loaded bool
}

func (c *ptrChild) Load(req *http.Request) error {
	c.Loaded = true
	return nil
}

type ptrParent struct {
	Child  *ptrChild
	Loaded bool
}

func (p *ptrParent) Load(req *http.Request) error {
	p.Loaded = true
	return nil
}

var errLoadFailed = errors.New("load failed")

type errLeaf struct{}

func (e *errLeaf) Load(req *http.Request) error { return errLoadFailed }

type errRoot struct {
	Leaf   errLeaf
	Loaded bool
}

func (r *errRoot) Load(req *http.Request) error {
	r.Loaded = true
	return nil
}

type cycleNode struct {
	Self   *cycleNode
	Loaded bool
}

func (c *cycleNode) Load(req *http.Request) error {
	c.Loaded = true
	return nil
}

var skippedChildCalled bool

type skippedChild struct{}

func (s *skippedChild) Load(req *http.Request) error {
	skippedChildCalled = true
	return nil
}

type skippedParent struct {
	child  skippedChild // unexported — loader must skip this field
	Loaded bool
}

func (p *skippedParent) Load(req *http.Request) error {
	p.Loaded = true
	return nil
}

type containerLeaf struct {
	Loaded bool
}

func (c *containerLeaf) Load(req *http.Request) error {
	c.Loaded = true
	return nil
}

type containerVM struct {
	Leaf containerLeaf
}

func TestLoad(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	t.Run("flat_struct_loads_itself", func(t *testing.T) {
		vm := &selfLoader{}
		require.NoError(t, torque.Load(req, vm))
		require.Equal(t, "loaded", vm.Value)
	})

	t.Run("parent_loader_fires_before_nested_value_loader", func(t *testing.T) {
		var order []string
		vm := &orderMid{
			order: &order,
			Leaf:  orderLeaf{order: &order},
		}
		require.NoError(t, torque.Load(req, vm))
		require.Equal(t, []string{"orderMid", "orderLeaf"}, order)
	})

	t.Run("nil_pointer_loader_field_is_allocated_and_loaded", func(t *testing.T) {
		vm := &ptrParent{Child: nil}
		require.NoError(t, torque.Load(req, vm))
		require.NotNil(t, vm.Child)
		require.True(t, vm.Child.Loaded)
		require.True(t, vm.Loaded)
	})

	t.Run("top_down_pre_order_across_three_levels", func(t *testing.T) {
		var order []string
		vm := &orderRoot{
			order: &order,
			Mid: orderMid{
				order: &order,
				Leaf:  orderLeaf{order: &order},
			},
		}
		require.NoError(t, torque.Load(req, vm))
		require.Equal(t, []string{"orderRoot", "orderMid", "orderLeaf"}, order)
	})

	t.Run("loader_error_propagates_wrapped_with_type_name", func(t *testing.T) {
		vm := &errRoot{}
		err := torque.Load(req, vm)
		require.Error(t, err)
		require.ErrorIs(t, err, errLoadFailed)
		require.True(t, strings.Contains(err.Error(), "errLeaf"))
		// parent runs before child with top-down loading, so it IS loaded when the child errors
		require.True(t, vm.Loaded)
	})

	t.Run("cyclic_pointer_field_does_not_recurse_infinitely", func(t *testing.T) {
		vm := &cycleNode{}
		require.NoError(t, torque.Load(req, vm))
		require.True(t, vm.Loaded)
		require.Nil(t, vm.Self)
	})

	t.Run("unexported_loader_fields_are_skipped", func(t *testing.T) {
		skippedChildCalled = false
		vm := &skippedParent{}
		require.NoError(t, torque.Load(req, vm))
		require.False(t, skippedChildCalled)
		require.True(t, vm.Loaded)
	})

	t.Run("non_loader_container_still_loads_its_loader_fields", func(t *testing.T) {
		vm := &containerVM{}
		require.NoError(t, torque.Load(req, vm))
		require.True(t, vm.Leaf.Loaded)
	})
}
