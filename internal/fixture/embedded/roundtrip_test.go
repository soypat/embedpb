// Package shapespb here is the embedpb-generated mirror of
// internal/fixture/stock, checked in by genpb's TestFixturesUpToDate. These
// tests are the generator's correctness suite: they pin the wire bytes, prove
// interop with the stock google.golang.org/protobuf runtime in both directions,
// and prove AppendJSON agrees with protojson.
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

// goldenHex is our canonical deterministic marshal of sample(), captured from a
// native run of protobuf-eval/path3b-shapes. It is NOT the stock runtime's byte
// order — wire order is free, and the stock runtime appends oneofs after other
// fields. Interop with the stock runtime is proven separately, below.
const goldenHex = "08f9ffffffffffffffff011080808080802018ac0220808080808080800228f1c00130d3db80cb493defbeadde41efcdab89674523014dc3f54840519b91048b0abf05405801620668c3a96c6c6f6a06000102fdfeff70027a0a082a12066e65737465648201070102ac02f0a2048a0105616c7068618a010462657461920105080112017892010508021201799a0109080912056f6e656f66aa010c0a026b311206080b12027631aa010c0a026b321206081612027632b00101ba011272656c61792e6d6f636b2e696e76616c6964c20104c000020a"

// sample builds a fully-populated message exercising every feature, including
// negative values (int32 varint sign-extension + zigzag) and a 2-key map.
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
		OptionalFlag: proto.Bool(true),
		OptionalName: proto.String("relay.mock.invalid"),
		OptionalBlob: []byte{192, 0, 2, 10},
	}
}

// stockSample is the same message built with the stock generated types.
func stockSample() *stock.Shapes {
	return &stock.Shapes{
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
		Kind:  stock.Kind_KIND_B,
		Inner: &stock.Inner{A: 42, B: "nested"},
		Nums:  []uint32{1, 2, 300, 70000},
		Names: []string{"alpha", "beta"},
		Inners: []*stock.Inner{
			{A: 1, B: "x"},
			{A: 2, B: "y"},
		},
		Choice: &stock.Shapes_ChoiceMsg{ChoiceMsg: &stock.Inner{A: 9, B: "oneof"}},
		Entries: map[string]*stock.Inner{
			"k1": {A: 11, B: "v1"},
			"k2": {A: 22, B: "v2"},
		},
		OptionalFlag: proto.Bool(true),
		OptionalName: proto.String("relay.mock.invalid"),
		OptionalBlob: []byte{192, 0, 2, 10},
	}
}

func golden(t *testing.T) []byte {
	t.Helper()
	b, err := hex.DecodeString(goldenHex)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

// TestMarshal exercises the whole reason the generator exists: proto.Marshal
// reaching our hand-written protoiface.Methods. Any reflective fallback panics
// in the msgReflect canaries instead of reaching the assertions.
func TestMarshal(t *testing.T) {
	want := golden(t)
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
}

func TestUnmarshalRoundTrip(t *testing.T) {
	want := golden(t)
	out := &Shapes{}
	if err := proto.Unmarshal(want, out); err != nil {
		t.Fatal(err)
	}
	if out.I32 != -7 || out.S64 != -9876543210 || out.Kind != Kind_KIND_B ||
		out.Inner == nil || out.Inner.B != "nested" || len(out.Nums) != 4 ||
		len(out.Names) != 2 || len(out.Inners) != 2 || len(out.Entries) != 2 ||
		out.Entries["k2"].B != "v2" || out.OptionalFlag == nil ||
		!*out.OptionalFlag || out.GetOptionalName() != "relay.mock.invalid" ||
		!bytes.Equal(out.GetOptionalBlob(), []byte{192, 0, 2, 10}) {
		t.Fatalf("decoded fields wrong: %+v", out)
	}
	cm, ok := out.Choice.(*Shapes_ChoiceMsg)
	if !ok || cm.ChoiceMsg.B != "oneof" {
		t.Fatalf("decoded oneof wrong: %#v", out.Choice)
	}
	re, err := proto.Marshal(out)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(re, want) {
		t.Fatalf("re-marshal mismatch\n got  %x\n want %x", re, want)
	}
}

// TestStockInterop proves the two codecs agree on the wire in both directions.
// The stock runtime is the reference; it is what every peer on the other end of
// the connection is running.
func TestStockInterop(t *testing.T) {
	ours, err := proto.Marshal(sample())
	if err != nil {
		t.Fatal(err)
	}
	st := &stock.Shapes{}
	if err := proto.Unmarshal(ours, st); err != nil {
		t.Fatalf("stock unmarshal of our bytes: %v", err)
	}
	if st.I32 != -7 || st.S64 != -9876543210 || st.Kind != stock.Kind_KIND_B ||
		st.Inner == nil || st.Inner.B != "nested" || len(st.Nums) != 4 ||
		len(st.Names) != 2 || len(st.Inners) != 2 || len(st.Entries) != 2 ||
		st.Entries["k2"].B != "v2" || st.OptionalFlag == nil ||
		!*st.OptionalFlag || st.GetOptionalName() != "relay.mock.invalid" ||
		!bytes.Equal(st.GetOptionalBlob(), []byte{192, 0, 2, 10}) {
		t.Fatalf("stock decoded our bytes wrong: %+v", st)
	}
	if cm, ok := st.Choice.(*stock.Shapes_ChoiceMsg); !ok || cm.ChoiceMsg.B != "oneof" {
		t.Fatalf("stock decoded our oneof wrong: %#v", st.Choice)
	}

	stockBytes, err := proto.Marshal(stockSample())
	if err != nil {
		t.Fatal(err)
	}
	mine := &Shapes{}
	if err := proto.Unmarshal(stockBytes, mine); err != nil {
		t.Fatalf("our unmarshal of stock bytes: %v", err)
	}
	if mine.I32 != -7 || mine.S64 != -9876543210 || mine.Kind != Kind_KIND_B ||
		mine.Inner == nil || mine.Inner.B != "nested" || len(mine.Nums) != 4 ||
		len(mine.Names) != 2 || len(mine.Inners) != 2 || len(mine.Entries) != 2 ||
		mine.Entries["k2"].B != "v2" || mine.OptionalFlag == nil ||
		!*mine.OptionalFlag || mine.GetOptionalName() != "relay.mock.invalid" ||
		!bytes.Equal(mine.GetOptionalBlob(), []byte{192, 0, 2, 10}) {
		t.Fatalf("we decoded stock bytes wrong: %+v", mine)
	}
	if cm, ok := mine.Choice.(*Shapes_ChoiceMsg); !ok || cm.ChoiceMsg.B != "oneof" {
		t.Fatalf("we decoded stock oneof wrong: %#v", mine.Choice)
	}
}

// TestAppendJSONMatchesProtojson pins AppendJSON to the protojson options
// netbird's wasm bridge uses. Compared as decoded values: protojson injects a
// random space (internal/detrand), so its bytes are deliberately unstable.
func TestAppendJSONMatchesProtojson(t *testing.T) {
	ours, err := sample().AppendJSON(nil)
	if err != nil {
		t.Fatal(err)
	}
	theirs, err := protojson.MarshalOptions{
		EmitUnpopulated: true,
		UseProtoNames:   true,
		AllowPartial:    true,
	}.Marshal(stockSample())
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

// TestMarshalSoak catches buffer-aliasing mistakes in the generated append
// paths, which show up as corruption only after reuse.
func TestMarshalSoak(t *testing.T) {
	want := golden(t)
	msg := sample()
	for i := range 1000 {
		b, err := proto.Marshal(msg)
		if err != nil {
			t.Fatalf("iteration %d: %v", i, err)
		}
		if !bytes.Equal(b, want) {
			t.Fatalf("iteration %d: wire drift %x", i, b)
		}
		if err := proto.Unmarshal(b, &Shapes{}); err != nil {
			t.Fatalf("iteration %d: %v", i, err)
		}
	}
}
