package gen

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// stampStock adds `//go:build !<tag>` to every stock *.pb.go in dir so the stock
// (reflection-based) files are excluded from the build the generated file
// targets. Idempotent: skips files already carrying the constraint, and skips
// our own generated output.
func stampStock(dir, tag string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	neg := "!" + tag
	want := "//go:build " + neg
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() || !strings.HasSuffix(name, ".pb.go") {
			continue
		}
		// Skip grpc service stubs: they define no messages (we don't redefine
		// their symbols) and must stay in BOTH builds — the reflection-free
		// build still needs the client stubs.
		if strings.HasSuffix(name, "_grpc.pb.go") {
			continue
		}
		path := filepath.Join(dir, name)
		src, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if bytes.Contains(src, []byte(want)) {
			continue // already stamped
		}
		stamped, err := insertBuildConstraint(src, want)
		if err != nil {
			return fmt.Errorf("%s: %w", name, err)
		}
		if err := os.WriteFile(path, stamped, 0o644); err != nil {
			return err
		}
		fmt.Printf("embedpb: stamped %s with %s\n", name, want)
	}
	return nil
}

// insertBuildConstraint puts the build-constraint line above the package clause,
// preserving any leading comment block (license header) and a blank line after
// the constraint (required by the go:build syntax).
func insertBuildConstraint(src []byte, line string) ([]byte, error) {
	lines := bytes.Split(src, []byte("\n"))
	pkgIdx := -1
	for i, l := range lines {
		if bytes.HasPrefix(bytes.TrimSpace(l), []byte("package ")) {
			pkgIdx = i
			break
		}
	}
	if pkgIdx < 0 {
		return nil, fmt.Errorf("no package clause found")
	}
	// Insert the constraint immediately before the package clause, with the
	// required trailing blank line.
	var out [][]byte
	out = append(out, lines[:pkgIdx]...)
	out = append(out, []byte(line), []byte(""))
	out = append(out, lines[pkgIdx:]...)
	return bytes.Join(out, []byte("\n")), nil
}
