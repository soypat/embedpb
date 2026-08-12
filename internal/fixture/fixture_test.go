package fixture

import (
	"os"
	"path/filepath"
	"testing"
)

// root is the module root relative to this package's directory, which is where
// go test runs.
const root = "../.."

// TestGeneratedFilesPresent reports the one thing `go test ./...` cannot fix
// for itself. The fixture packages import generated code, and a test cannot
// generate code its sibling packages were already built against within the same
// invocation. This package imports no generated code, so it always compiles and
// can say what to run.
func TestGeneratedFilesPresent(t *testing.T) {
	if testing.Short() {
		t.Skip("-short: assuming generated fixtures are present")
	}
	for _, f := range All {
		for _, out := range f.Outputs() {
			if _, err := os.Stat(filepath.Join(root, filepath.FromSlash(out))); err != nil {
				t.Errorf("missing generated fixture %s — run: go generate ./...", out)
			}
		}
	}
}
