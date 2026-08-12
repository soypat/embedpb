// Package shapespb here is the embedpb-generated mirror of
// internal/fixture/stock, produced by `go generate ./...`. These tests are the
// generator's correctness suite: fixtest proves interop with the stock
// google.golang.org/protobuf runtime in both directions and AppendJSON parity
// with protojson; the samples below are what gets put through it.
package shapespb

import (
	"bytes"
	"testing"
	"time"

	"github.com/soypat/embedpb/internal/fixture/fixtest"
	stock "github.com/soypat/embedpb/internal/fixture/stock"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/durationpb"
)

// sample is fully populated, including negative values (int32 varint
// sign-extension + zigzag) and a 2-key map.
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
		OptionalFlag: new(true),
		OptionalName: new("relay.mock.invalid"),
		OptionalBlob: []byte{192, 0, 2, 10},
		Labels: map[string]string{
			"env":  "test",
			"role": "proxy",
		},
		OptionalI32:      new(int32(-321)),
		OptionalI64:      new(int64(51820)),
		OptionalU32:      new(uint32(8)),
		OptionalDuration: durationpb.New(90 * time.Second),
		Resolved: map[string]*StringList{
			"peer-a": {Values: []string{"10.0.0.10", "fd00::10"}},
			"peer-b": {Values: []string{"10.0.0.20"}},
		},
		Auth:          &Shapes_Header{Header: &AuthHeader{Name: "x-user", Value: "alice"}},
		Kinds:         []Kind{Kind_KIND_A, Kind_KIND_B},
		Blobs:         [][]byte{[]byte("alpha"), []byte("beta")},
		OptionalKind:  Kind_KIND_B.Enum(),
		OptionalInner: &Inner{A: 77, B: "optional-local"},
		ScalarChoice:  &Shapes_ChoiceText{ChoiceText: "scalar-oneof"},
	}
}

// optionalZeroValues sets every optional field to its type's zero value. Set is
// not unset: each one must survive the round trip as present. Optional bytes is
// the case that caught the generator emitting append([]byte(nil), v...).
func optionalZeroValues() *Shapes {
	return &Shapes{
		OptionalFlag:     new(false),
		OptionalName:     new(""),
		OptionalBlob:     []byte{},
		OptionalI32:      new(int32(0)),
		OptionalI64:      new(int64(0)),
		OptionalU32:      new(uint32(0)),
		OptionalDuration: durationpb.New(0),
		OptionalKind:     Kind_KIND_UNKNOWN.Enum(),
		OptionalInner:    &Inner{},
	}
}

func TestShapes(t *testing.T) {
	fixtest.Run(t, []fixtest.Case{
		{Name: "full", Ours: sample(), Stock: new(stock.Shapes)},
		{Name: "empty", Ours: &Shapes{}, Stock: new(stock.Shapes)},
		{Name: "optional-zero-values-set", Ours: optionalZeroValues(), Stock: new(stock.Shapes)},

		// Message-armed oneof.
		{Name: "auth-password", Ours: &Shapes{Auth: &Shapes_Password{Password: &AuthPassword{Password: "secret"}}}, Stock: new(stock.Shapes)},
		{Name: "auth-pin", Ours: &Shapes{Auth: &Shapes_Pin{Pin: &AuthPin{Pin: "123456"}}}, Stock: new(stock.Shapes)},
		{Name: "auth-header", Ours: &Shapes{Auth: &Shapes_Header{Header: &AuthHeader{Name: "x-user", Value: "alice"}}}, Stock: new(stock.Shapes)},

		// Scalar-armed oneof.
		{Name: "choice-text", Ours: &Shapes{ScalarChoice: &Shapes_ChoiceText{ChoiceText: "hello"}}, Stock: new(stock.Shapes)},
		{Name: "choice-data", Ours: &Shapes{ScalarChoice: &Shapes_ChoiceData{ChoiceData: []byte{1, 2, 3}}}, Stock: new(stock.Shapes)},
		{Name: "choice-flag", Ours: &Shapes{ScalarChoice: &Shapes_ChoiceFlag{ChoiceFlag: true}}, Stock: new(stock.Shapes)},
	})
}

// TestMarshalSoak catches buffer-aliasing mistakes in the generated append
// paths, which show up as corruption only after reuse. The reference is the
// first marshal's own output, derived here rather than captured.
func TestMarshalSoak(t *testing.T) {
	msg := sample()
	want, err := proto.Marshal(msg)
	if err != nil {
		t.Fatal(err)
	}
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
