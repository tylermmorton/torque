package torque_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
	torque2 "github.com/tylermmorton/torque"
)

type ctxKey string

const testCtxKey ctxKey = "test"

// --- flat struct ---

type selfContextProvider struct{}

func (s *selfContextProvider) Context(req *http.Request) *http.Request {
	return torque2.Provide(req, testCtxKey, "self")
}

// --- three-level tree for ordering ---

type ctxOrderLeaf struct {
	order *[]string
}

func (l *ctxOrderLeaf) Context(req *http.Request) *http.Request {
	*l.order = append(*l.order, "leaf")
	return req
}

type ctxOrderMid struct {
	order *[]string
	Leaf  ctxOrderLeaf
}

func (m *ctxOrderMid) Context(req *http.Request) *http.Request {
	*m.order = append(*m.order, "mid")
	return req
}

type ctxOrderRoot struct {
	order *[]string
	Mid   ctxOrderMid
}

func (r *ctxOrderRoot) Context(req *http.Request) *http.Request {
	*r.order = append(*r.order, "root")
	return req
}

// --- subtree threading siblings ---

type ctxSiblingA struct {
	// Captures what value it received for testCtxKey, then overrides it.
	Saw string
}

func (a *ctxSiblingA) Context(req *http.Request) *http.Request {
	val, _ := torque2.Inject[string](req, testCtxKey)
	a.Saw = val
	return torque2.Provide(req, testCtxKey, "childA")
}

type ctxSiblingB struct {
	// Captures what value it received for testCtxKey without modifying it.
	Saw string
}

func (b *ctxSiblingB) Context(req *http.Request) *http.Request {
	val, _ := torque2.Inject[string](req, testCtxKey)
	b.Saw = val
	return req
}

type ctxSiblingRoot struct {
	ChildA ctxSiblingA
	ChildB ctxSiblingB
}

func (r *ctxSiblingRoot) Context(req *http.Request) *http.Request {
	return torque2.Provide(req, testCtxKey, "root")
}

// --- handler-level: context before loader ---

type ctxBeforeLoaderKey struct{}

type ctxBeforeLoaderVM struct {
	ContextRanFirst bool `json:"context_ran_first"`
}

func (v *ctxBeforeLoaderVM) Context(req *http.Request) *http.Request {
	return torque2.Provide(req, ctxBeforeLoaderKey{}, true)
}

func (v *ctxBeforeLoaderVM) Load(req *http.Request) error {
	v.ContextRanFirst, _ = torque2.Inject[bool](req, ctxBeforeLoaderKey{})
	return nil
}

// --- handler-level: context before action ---

type ctxBeforeActionVM struct{}

func (v *ctxBeforeActionVM) Context(req *http.Request) *http.Request {
	return torque2.Provide(req, ctxBeforeLoaderKey{}, true)
}

func (v *ctxBeforeActionVM) Action(wr http.ResponseWriter, req *http.Request) error {
	ran, _ := torque2.Inject[bool](req, ctxBeforeLoaderKey{})
	if ran {
		wr.WriteHeader(http.StatusOK)
	} else {
		wr.WriteHeader(http.StatusInternalServerError)
	}
	return nil
}

// ---

func TestContext(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	t.Run("flat_struct_calls_context_on_itself", func(t *testing.T) {
		vm := &selfContextProvider{}
		result := torque2.Context(req, vm)
		val, ok := torque2.Inject[string](result, testCtxKey)
		require.True(t, ok)
		require.Equal(t, "self", val)
	})

	t.Run("pre_order_dfs_across_three_levels", func(t *testing.T) {
		var order []string
		vm := &ctxOrderRoot{
			order: &order,
			Mid: ctxOrderMid{
				order: &order,
				Leaf:  ctxOrderLeaf{order: &order},
			},
		}
		torque2.Context(req, vm)
		require.Equal(t, []string{"root", "mid", "leaf"}, order)
	})

	t.Run("subtree_threading_siblings_see_parent_req_not_each_other", func(t *testing.T) {
		vm := &ctxSiblingRoot{}
		torque2.Context(req, vm)
		// Both children should see "root" (what the parent set), not each other's value.
		require.Equal(t, "root", vm.ChildA.Saw)
		require.Equal(t, "root", vm.ChildB.Saw)
	})

	t.Run("context_runs_on_get_before_loader", func(t *testing.T) {
		r := torque2.NewRouter()
		r.Handle("/", torque2.MustNewHandler[ctxBeforeLoaderVM]())

		getReq := httptest.NewRequest(http.MethodGet, "/", nil)
		getReq.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, getReq)

		require.Equal(t, http.StatusOK, rec.Code)
		require.Contains(t, rec.Body.String(), `"context_ran_first":true`)
	})

	for _, method := range []string{
		http.MethodPost,
		http.MethodPut,
		http.MethodPatch,
		http.MethodDelete,
	} {
		t.Run("context_runs_on_"+method+"_before_action", func(t *testing.T) {
			r := torque2.NewRouter()
			r.Handle("/", torque2.MustNewHandler[ctxBeforeActionVM]())

			actionReq := httptest.NewRequest(method, "/", nil)
			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, actionReq)

			require.Equal(t, http.StatusOK, rec.Code)
		})
	}
}
