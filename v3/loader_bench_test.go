package torque

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
)

type benchLeaf struct{ Loaded bool }

func (b *benchLeaf) Load(req *http.Request) error { b.Loaded = true; return nil }

type benchMid struct {
	Leaf   benchLeaf
	Loaded bool
}

func (b *benchMid) Load(req *http.Request) error { b.Loaded = true; return nil }

type benchRoot struct {
	Mid    benchMid
	Loaded bool
}

func (b *benchRoot) Load(req *http.Request) error { b.Loaded = true; return nil }

// BenchmarkLoad_ColdPlan measures plan compilation cost by clearing the cache
// before each iteration.
func BenchmarkLoad_ColdPlan(b *testing.B) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		planCache = sync.Map{}
		vm := &benchRoot{}
		_ = Load(req, vm)
	}
}

// BenchmarkLoad_WarmPlan measures steady-state execution cost with a hot plan cache.
func BenchmarkLoad_WarmPlan(b *testing.B) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	_ = Load(req, &benchRoot{})
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		vm := &benchRoot{}
		_ = Load(req, vm)
	}
}
