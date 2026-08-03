package gen

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// stampStock is what lets the stock and generated files coexist in one package,
// so the deployment recipe (commit the stamps, regenerate the companion at
// build time) depends on it being idempotent and on it leaving service stubs
// alone.
func TestStampStock(t *testing.T) {
	const licensed = `// Copyright 2020 The Authors.
// SPDX-License-Identifier: Apache-2.0

package proto

type Msg struct{}
`
	const grpcStub = `package proto

type Client interface{}
`
	dir := t.TempDir()
	write := func(name, src string) string {
		p := filepath.Join(dir, name)
		if err := os.WriteFile(p, []byte(src), 0o644); err != nil {
			t.Fatal(err)
		}
		return p
	}
	msgFile := write("thing.pb.go", licensed)
	grpcFile := write("thing_grpc.pb.go", grpcStub)
	other := write("hand_written.go", "package proto\n")

	if err := stampStock(dir, "tinygo"); err != nil {
		t.Fatal(err)
	}
	got := read(t, msgFile)
	if !strings.Contains(got, "//go:build !tinygo\n\npackage proto") {
		t.Fatalf("constraint not inserted above the package clause:\n%s", got)
	}
	if !strings.HasPrefix(got, "// Copyright") {
		t.Fatalf("license header not preserved:\n%s", got)
	}
	if s := read(t, grpcFile); s != grpcStub {
		t.Errorf("service stub was stamped; it belongs to both builds:\n%s", s)
	}
	if s := read(t, other); s != "package proto\n" {
		t.Errorf("non-.pb.go file was stamped:\n%s", s)
	}

	// Idempotent: regeneration re-runs this on already-stamped files.
	if err := stampStock(dir, "tinygo"); err != nil {
		t.Fatal(err)
	}
	if s := read(t, msgFile); s != got {
		t.Errorf("second stamp changed the file:\n%s", s)
	}
}

func TestInsertBuildConstraintNoPackage(t *testing.T) {
	if _, err := insertBuildConstraint([]byte("// just a comment\n"), "//go:build !tinygo"); err == nil {
		t.Fatal("expected an error when there is no package clause")
	}
}

func read(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}
