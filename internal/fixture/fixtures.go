// Package fixture is the single source of truth for the generator's test
// fixtures: which stock package each one reads, and where its generated output
// lands. It imports no generated code, so it always compiles.
//
// Paths are slash-separated and relative to the module root; each consumer
// joins them onto its own root.
package fixture

// Fixture is one stock package and the generated files made from it.
type Fixture struct {
	Name      string
	StockDir  string
	InlineOut string
	// RuntimeOut is "" when the fixture has no compiled runtime package.
	// Runtime mode is still generator-checked for every fixture; see
	// genpb's TestRuntimeModeOnlySwapsSupportCode.
	RuntimeOut string
}

// Outputs returns the generated files this fixture owns.
func (f Fixture) Outputs() []string {
	if f.RuntimeOut == "" {
		return []string{f.InlineOut}
	}
	return []string{f.InlineOut, f.RuntimeOut}
}

// All is every fixture, in generation order.
//
// shapes is hand-written to cover every wire shape. daemon and management-net
// are real netbird protos kept for constructs shapes cannot express without
// protoc: nested message and enum declarations, and a five-level message chain.
var All = []Fixture{
	{
		Name:       "shapes",
		StockDir:   "internal/fixture/stock",
		InlineOut:  "internal/fixture/embedded/embedpb_generated.go",
		RuntimeOut: "internal/fixture/embeddedrt/embedpb_generated.go",
	},
	{
		Name:      "management-net",
		StockDir:  "internal/fixture/managementnetstock",
		InlineOut: "internal/fixture/managementnetembedded/embedpb_generated.go",
	},
	{
		Name:      "daemon",
		StockDir:  "internal/fixture/daemonstock",
		InlineOut: "internal/fixture/daemonembedded/embedpb_generated.go",
	},
}
