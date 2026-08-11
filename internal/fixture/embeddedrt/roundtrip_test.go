// Package shapespb here is the same fixture as internal/fixture/embedded,
// generated with Options.Runtime — the support code comes from
// github.com/soypat/embedpb/pbruntime instead of being emitted inline.
//
// genpb's TestRuntimeModeOnlySwapsSupportCode already proves the two modes emit
// the same message bodies, for every fixture. This package is the other half:
// the one compiled runtime fixture, proving the aliases point at code that
// actually runs.
package shapespb

import (
	"testing"

	"github.com/soypat/embedpb/internal/fixture/fixtest"
	stock "github.com/soypat/embedpb/internal/fixture/stock"
)

func sample() *Shapes {
	return &Shapes{
		I32:   -7,
		I64:   1 << 40,
		U32:   300,
		U64:   1 << 50,
		S32:   -12345,
		S64:   -9876543210,
		F32:   0xDEADBEEF,
		F64:   0x0123456789ABCDEF,
		Fl:    3.14,
		Db:    2.718281828,
		Flag:  true,
		Str:   "héllo",
		Blob:  []byte{0, 1, 2, 253, 254, 255},
		Kind:  Kind_KIND_B,
		Inner: &Inner{A: 42, B: "nested"},
		Nums:  []uint32{1, 2, 300, 70000},
		Names: []string{"alpha", "beta"},
		Inners: []*Inner{
			{A: 1, B: "x"},
			{A: 2, B: "y"},
		},
		Choice: &Shapes_ChoiceMsg{ChoiceMsg: &Inner{A: 9, B: "oneof"}},
		Entries: map[string]*Inner{
			"k1": {A: 11, B: "v1"},
			"k2": {A: 22, B: "v2"},
		},
	}
}

func TestRuntimeShim(t *testing.T) {
	fixtest.Run(t, []fixtest.Case{
		{Name: "populated", Ours: sample(), Stock: new(stock.Shapes)},
		{Name: "empty", Ours: &Shapes{}, Stock: new(stock.Shapes)},
	})
}
