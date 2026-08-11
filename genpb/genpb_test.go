package genpb_test

import (
	"bytes"
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/soypat/embedpb/genpb"
	"github.com/soypat/embedpb/internal/fixture"
)

var update = flag.Bool("update", false, "rewrite the checked-in fixtures with the current output")

// root is the module root relative to this package's directory, which is where
// go test runs.
const root = ".."

const stockDir = "../internal/fixture/stock"

// TestFixturesUpToDate compares regenerated output against what is on disk.
//
// The generated files are gitignored and CI regenerates before testing, so in
// CI this check is a no-op by construction. Its value is local: it catches a
// working tree that has been tested more often than it has been regenerated.
func TestFixturesUpToDate(t *testing.T) {
	for _, f := range fixture.All {
		for _, m := range modes(f) {
			t.Run(f.Name+"-"+m.suffix, func(t *testing.T) {
				got := generate(t, f.StockDir, m.runtime)
				golden := filepath.Join(root, filepath.FromSlash(m.out))
				if *update {
					if err := os.WriteFile(golden, got, 0o644); err != nil {
						t.Fatal(err)
					}
					t.Logf("wrote %s (%d bytes)", m.out, len(got))
					return
				}
				want, err := os.ReadFile(golden)
				if err != nil {
					t.Fatal(err)
				}
				if !bytes.Equal(got, want) {
					t.Errorf("%s is stale; regenerate with:\n\tgo generate ./...", m.out)
				}
			})
		}
	}
}

// Support-code boundaries. Everything from these markers on is the shim, which
// is the half of the output -runtime is allowed to change.
const (
	inlineShimStart  = "\nfunc embedBool("
	runtimeShimStart = "\ntype msgReflect = "

	protoifaceImport = `"google.golang.org/protobuf/runtime/protoiface"`
	pbruntimeImport  = `"github.com/soypat/embedpb/pbruntime"`
)

// TestRuntimeModeOnlySwapsSupportCode is what licenses having one compiled
// runtime fixture instead of one per proto: -runtime swaps the shim and the
// import that reaches it, and touches nothing else. Checked for every fixture,
// for the cost of generating it twice.
func TestRuntimeModeOnlySwapsSupportCode(t *testing.T) {
	for _, f := range fixture.All {
		t.Run(f.Name, func(t *testing.T) {
			inline := messageBody(t, generate(t, f.StockDir, false), inlineShimStart)
			runtime := messageBody(t, generate(t, f.StockDir, true), runtimeShimStart)
			if inline != runtime {
				t.Errorf("-runtime changed more than the support code; first diff at byte %d", firstDiff(inline, runtime))
			}
		})
	}
}

// messageBody strips the two things -runtime is expected to change: the support
// code, which starts at marker, and the import line that reaches it.
func messageBody(t *testing.T, src []byte, marker string) string {
	t.Helper()
	body, _, ok := strings.Cut(string(src), marker)
	if !ok {
		t.Fatalf("support-code marker %q not in generated output", marker)
	}
	var b strings.Builder
	for line := range strings.Lines(body) {
		switch strings.TrimSpace(line) {
		case protoifaceImport, pbruntimeImport:
			continue
		}
		b.WriteString(line)
	}
	return b.String()
}

func firstDiff(a, b string) int {
	for i := 0; i < min(len(a), len(b)); i++ {
		if a[i] != b[i] {
			return i
		}
	}
	return min(len(a), len(b))
}

type mode struct {
	suffix  string
	out     string
	runtime bool
}

func modes(f fixture.Fixture) []mode {
	out := []mode{{suffix: "inline", out: f.InlineOut}}
	if f.RuntimeOut != "" {
		out = append(out, mode{suffix: "runtime", out: f.RuntimeOut, runtime: true})
	}
	return out
}

func generate(t *testing.T, stockDir string, runtime bool) []byte {
	t.Helper()
	out := filepath.Join(t.TempDir(), "embedpb_generated.go")
	err := genpb.Generate(genpb.Options{
		Dir:     filepath.Join(root, filepath.FromSlash(stockDir)),
		Out:     out,
		Runtime: runtime,
		// No Tag: the fixtures must build under a plain `go test ./...`.
		// StampStock stays off; the stock package is the input, it must
		// keep building too.
	})
	if err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

// TestStampNeedsTag guards the one Options combination that cannot mean
// anything: there is no negation to stamp without a tag.
func TestStampNeedsTag(t *testing.T) {
	err := genpb.Generate(genpb.Options{Dir: stockDir, StampStock: true, Out: filepath.Join(t.TempDir(), "x.go")})
	if err == nil {
		t.Fatal("expected an error for StampStock without Tag")
	}
}
