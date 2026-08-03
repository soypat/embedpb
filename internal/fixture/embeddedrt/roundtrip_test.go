// Package shapespb here is the same fixture as internal/fixture/embedded,
// generated with Options.Runtime — the support code comes from
// github.com/soypat/embedpb/pbruntime instead of being emitted inline. The
// point of these tests is that swapping the shim changes nothing observable:
// same wire bytes, same interop, same JSON.
package shapespb

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"reflect"
	"testing"

	stock "github.com/soypat/embedpb/internal/fixture/stock"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

// goldenHex is the same constant the inline fixture pins; identical output from
// both modes is the assertion.
const goldenHex = "08f9ffffffffffffffff011080808080802018ac0220808080808080800228f1c00130d3db80cb493defbeadde41efcdab89674523014dc3f54840519b91048b0abf05405801620668c3a96c6c6f6a06000102fdfeff70027a0a082a12066e65737465648201070102ac02f0a2048a0105616c7068618a010462657461920105080112017892010508021201799a0109080912056f6e656f66aa010c0a026b311206080b12027631aa010c0a026b321206081612027632"

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

func TestRuntimeShimMarshal(t *testing.T) {
	want, err := hex.DecodeString(goldenHex)
	if err != nil {
		t.Fatal(err)
	}
	got, err := proto.Marshal(sample())
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, want) {
		t.Fatalf("wire mismatch\n got  %x\n want %x", got, want)
	}
	if n := proto.Size(sample()); n != len(want) {
		t.Errorf("proto.Size = %d, want %d", n, len(want))
	}
	out := &Shapes{}
	if err := proto.Unmarshal(want, out); err != nil {
		t.Fatal(err)
	}
	if out.I32 != -7 || out.Kind != Kind_KIND_B || out.Entries["k2"].B != "v2" {
		t.Fatalf("decoded fields wrong: %+v", out)
	}
}

// TestRuntimeShimJSON checks the pbruntime JSON helpers are wired up: the
// aliases are the only place the two modes differ in the JSON path.
func TestRuntimeShimJSON(t *testing.T) {
	ours, err := sample().AppendJSON(nil)
	if err != nil {
		t.Fatal(err)
	}
	st := &stock.Shapes{}
	if err := proto.Unmarshal(mustMarshal(t, sample()), st); err != nil {
		t.Fatal(err)
	}
	theirs, err := protojson.MarshalOptions{
		EmitUnpopulated: true,
		UseProtoNames:   true,
		AllowPartial:    true,
	}.Marshal(st)
	if err != nil {
		t.Fatal(err)
	}
	var gotV, wantV any
	if err := json.Unmarshal(ours, &gotV); err != nil {
		t.Fatalf("our JSON is invalid: %v\n%s", err, ours)
	}
	if err := json.Unmarshal(theirs, &wantV); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(gotV, wantV) {
		t.Errorf("JSON mismatch\n got  %s\n want %s", ours, theirs)
	}
}

func mustMarshal(t *testing.T, m proto.Message) []byte {
	t.Helper()
	b, err := proto.Marshal(m)
	if err != nil {
		t.Fatal(err)
	}
	return b
}
