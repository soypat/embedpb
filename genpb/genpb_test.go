package genpb_test

import (
	"bytes"
	"flag"
	"os"
	"path/filepath"
	"testing"

	"github.com/soypat/embedpb/genpb"
)

var update = flag.Bool("update", false, "rewrite the checked-in fixtures with the current output")

const (
	stockDir    = "../internal/fixture/stock"
	embeddedOut = "../internal/fixture/embedded/embedpb_generated.go"
	runtimeOut  = "../internal/fixture/embeddedrt/embedpb_generated.go"

	proxyStockDir    = "../internal/fixture/proxystock"
	proxyEmbeddedOut = "../internal/fixture/proxyembedded/embedpb_generated.go"
	proxyRuntimeOut  = "../internal/fixture/proxyembeddedrt/embedpb_generated.go"

	flowStockDir    = "../internal/fixture/flowstock"
	flowEmbeddedOut = "../internal/fixture/flowembedded/embedpb_generated.go"
	flowRuntimeOut  = "../internal/fixture/flowembeddedrt/embedpb_generated.go"

	managementJobStockDir    = "../internal/fixture/managementjobstock"
	managementJobEmbeddedOut = "../internal/fixture/managementjobembedded/embedpb_generated.go"
	managementJobRuntimeOut  = "../internal/fixture/managementjobembeddedrt/embedpb_generated.go"

	managementNetStockDir    = "../internal/fixture/managementnetstock"
	managementNetEmbeddedOut = "../internal/fixture/managementnetembedded/embedpb_generated.go"
	managementNetRuntimeOut  = "../internal/fixture/managementnetembeddedrt/embedpb_generated.go"
)

// TestFixturesUpToDate regenerates from the stock shapes package and compares
// against the checked-in output. The fixture packages are compiled and
// round-tripped by their own tests; this is what proves those packages are the
// current generator's output rather than a stale snapshot.
func TestFixturesUpToDate(t *testing.T) {
	for _, tt := range []struct {
		name    string
		dir     string
		runtime bool
		golden  string
	}{
		{name: "shapes-inline", dir: stockDir, golden: embeddedOut},
		{name: "shapes-runtime", dir: stockDir, runtime: true, golden: runtimeOut},
		{name: "proxy-inline", dir: proxyStockDir, golden: proxyEmbeddedOut},
		{name: "proxy-runtime", dir: proxyStockDir, runtime: true, golden: proxyRuntimeOut},
		{name: "flow-inline", dir: flowStockDir, golden: flowEmbeddedOut},
		{name: "flow-runtime", dir: flowStockDir, runtime: true, golden: flowRuntimeOut},
		{name: "management-job-inline", dir: managementJobStockDir, golden: managementJobEmbeddedOut},
		{name: "management-job-runtime", dir: managementJobStockDir, runtime: true, golden: managementJobRuntimeOut},
		{name: "management-net-inline", dir: managementNetStockDir, golden: managementNetEmbeddedOut},
		{name: "management-net-runtime", dir: managementNetStockDir, runtime: true, golden: managementNetRuntimeOut},
	} {
		t.Run(tt.name, func(t *testing.T) {
			out := filepath.Join(t.TempDir(), "embedpb_generated.go")
			err := genpb.Generate(genpb.Options{
				Dir:     tt.dir,
				Out:     out,
				Runtime: tt.runtime,
				// No Tag: the fixtures must build under a plain `go test ./...`.
				// StampStock stays off; the stock package is the input, it must
				// keep building too.
			})
			if err != nil {
				t.Fatal(err)
			}
			got, err := os.ReadFile(out)
			if err != nil {
				t.Fatal(err)
			}
			if *update {
				if err := os.WriteFile(tt.golden, got, 0o644); err != nil {
					t.Fatal(err)
				}
				t.Logf("wrote %s (%d bytes)", tt.golden, len(got))
				return
			}
			want, err := os.ReadFile(tt.golden)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(got, want) {
				t.Errorf("%s is stale; regenerate with:\n\tgo test ./genpb -run TestFixturesUpToDate -update", tt.golden)
			}
		})
	}
}

// TestStampNeedsTag guards the one Options combination that cannot mean
// anything: there is no negation to stamp without a tag.
func TestStampNeedsTag(t *testing.T) {
	err := genpb.Generate(genpb.Options{Dir: stockDir, StampStock: true, Out: filepath.Join(t.TempDir(), "x.go")})
	if err == nil {
		t.Fatal("expected an error for StampStock without Tag")
	}
}
