package torque_test

import (
	"io"
	"testing"

	torque "github.com/tylermmorton/torque/v3"
)

func BenchmarkCompileTemplate_L1(b *testing.B) {
	b.ReportAllocs()
	for range b.N {
		_, _ = torque.CompileTemplate(&tmplL1{})
	}
}

func BenchmarkCompileTemplate_L3(b *testing.B) {
	b.ReportAllocs()
	for range b.N {
		_, _ = torque.CompileTemplate(&tmplL3Root{})
	}
}

func BenchmarkCompileTemplate_L5(b *testing.B) {
	b.ReportAllocs()
	for range b.N {
		_, _ = torque.CompileTemplate(&tmplL5Root{})
	}
}

func BenchmarkRender_L1(b *testing.B) {
	tmpl, err := torque.CompileTemplate(&tmplL1{})
	if err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		_ = tmpl.Render(io.Discard, nil)
	}
}

func BenchmarkRender_L3(b *testing.B) {
	tmpl, err := torque.CompileTemplate(&tmplL3Root{})
	if err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		_ = tmpl.Render(io.Discard, nil)
	}
}

func BenchmarkRender_L5(b *testing.B) {
	tmpl, err := torque.CompileTemplate(&tmplL5Root{})
	if err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		_ = tmpl.Render(io.Discard, nil)
	}
}

func BenchmarkRender_L1_Parallel(b *testing.B) {
	tmpl, err := torque.CompileTemplate(&tmplL1{})
	if err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = tmpl.Render(io.Discard, nil)
		}
	})
}

func BenchmarkRender_L3_Parallel(b *testing.B) {
	tmpl, err := torque.CompileTemplate(&tmplL3Root{})
	if err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = tmpl.Render(io.Discard, nil)
		}
	})
}

func BenchmarkRender_L5_Parallel(b *testing.B) {
	tmpl, err := torque.CompileTemplate(&tmplL5Root{})
	if err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = tmpl.Render(io.Discard, nil)
		}
	})
}
