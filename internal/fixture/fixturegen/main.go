// Command fixturegen regenerates every fixture in fixture.All.
//
// It exists because cmd/embedpb requires -tag and the fixtures must build
// untagged, under a plain `go test ./...`.
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/soypat/embedpb/genpb"
	"github.com/soypat/embedpb/internal/fixture"
)

func main() {
	// go:generate runs in internal/fixture, two levels below the module root.
	root := flag.String("root", filepath.Join("..", ".."), "module root")
	flag.Parse()

	for _, f := range fixture.All {
		if err := generate(*root, f.StockDir, f.InlineOut, false); err != nil {
			fail(f.Name, err)
		}
		if f.RuntimeOut == "" {
			continue
		}
		if err := generate(*root, f.StockDir, f.RuntimeOut, true); err != nil {
			fail(f.Name, err)
		}
	}
}

func generate(root, stockDir, out string, runtime bool) error {
	return genpb.Generate(genpb.Options{
		Dir:     filepath.Join(root, filepath.FromSlash(stockDir)),
		Out:     filepath.Join(root, filepath.FromSlash(out)),
		Runtime: runtime,
		// No Tag, so no StampStock: the stock package is the input and must
		// keep building alongside the generated file.
	})
}

func fail(name string, err error) {
	fmt.Fprintf(os.Stderr, "fixturegen: %s: %v\n", name, err)
	os.Exit(1)
}
