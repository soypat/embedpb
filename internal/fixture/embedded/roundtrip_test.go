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
	"time"

	stock "github.com/soypat/embedpb/internal/fixture/stock"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/durationpb"
)

// goldenHex is our canonical deterministic marshal of sample(), captured from a
// native run of protobuf-eval/path3b-shapes. It is NOT the stock runtime's byte
// order — wire order is free, and the stock runtime appends oneofs after other
// fields. Interop with the stock runtime is proven separately, below.
const goldenHex = "08f9ffffffffffffffff011080808080802018ac0220808080808080800228f1c00130d3db80cb493defbeadde41efcdab89674523014dc3f54840519b91048b0abf05405801620668c3a96c6c6f6a06000102fdfeff70027a0a082a12066e65737465648201070102ac02f0a2048a0105616c7068618a010462657461920105080112017892010508021201799a0109080912056f6e656f66aa010c0a026b311206080b12027631aa010c0a026b321206081612027632b00101ba011272656c61792e6d6f636b2e696e76616c6964c20104c000020aca010b0a03656e76120474657374ca010d0a04726f6c65120570726f7879d001bffdffffffffffffff01d801ec9403e00108ea0102085af2011f0a06706565722d6112150a0931302e302e302e31300a08666430303a3a3130f201150a06706565722d62120b0a0931302e302e302e32308a020f0a06782d757365721205616c69636592020201029a0205616c7068619a020462657461a00202aa0212084d120e6f7074696f6e616c2d6c6f63616cb2020c7363616c61722d6f6e656f66"

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
		Labels: map[string]string{
			"env":  "test",
			"role": "proxy",
		},
		OptionalI32:      proto.Int32(-321),
		OptionalI64:      proto.Int64(51820),
		OptionalU32:      proto.Uint32(8),
		OptionalDuration: durationpb.New(90 * time.Second),
		Resolved: map[string]*StringList{
			"peer-a": {Values: []string{"10.0.0.10", "fd00::10"}},
			"peer-b": {Values: []string{"10.0.0.20"}},
		},
		Auth:           &Shapes_Header{Header: &AuthHeader{Name: "x-user", Value: "alice"}},
		Kinds:          []Kind{Kind_KIND_A, Kind_KIND_B},
		Blobs:          [][]byte{[]byte("alpha"), []byte("beta")},
		OptionalKind:   Kind_KIND_B.Enum(),
		OptionalInner:  &Inner{A: 77, B: "optional-local"},
		ScalarChoice:   &Shapes_ChoiceText{ChoiceText: "scalar-oneof"},
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
		Labels: map[string]string{
			"env":  "test",
			"role": "proxy",
		},
		OptionalI32:      proto.Int32(-321),
		OptionalI64:      proto.Int64(51820),
		OptionalU32:      proto.Uint32(8),
		OptionalDuration: durationpb.New(90 * time.Second),
		Resolved: map[string]*stock.StringList{
			"peer-a": {Values: []string{"10.0.0.10", "fd00::10"}},
			"peer-b": {Values: []string{"10.0.0.20"}},
		},
		Auth:           &stock.Shapes_Header{Header: &stock.AuthHeader{Name: "x-user", Value: "alice"}},
		Kinds:          []stock.Kind{stock.Kind_KIND_A, stock.Kind_KIND_B},
		Blobs:          [][]byte{[]byte("alpha"), []byte("beta")},
		OptionalKind:   stock.Kind_KIND_B.Enum(),
		OptionalInner:  &stock.Inner{A: 77, B: "optional-local"},
		ScalarChoice:   &stock.Shapes_ChoiceText{ChoiceText: "scalar-oneof"},
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
		!bytes.Equal(out.GetOptionalBlob(), []byte{192, 0, 2, 10}) ||
		len(out.Labels) != 2 || out.Labels["role"] != "proxy" ||
		out.OptionalI32 == nil || *out.OptionalI32 != -321 ||
		out.OptionalI64 == nil || *out.OptionalI64 != 51820 ||
		out.OptionalU32 == nil || *out.OptionalU32 != 8 ||
		out.OptionalDuration == nil || out.OptionalDuration.Seconds != 90 ||
		len(out.Resolved) != 2 || len(out.Resolved["peer-a"].Values) != 2 ||
		out.Resolved["peer-a"].Values[1] != "fd00::10" ||
		len(out.Kinds) != 2 || out.Kinds[1] != Kind_KIND_B ||
		len(out.Blobs) != 2 || !bytes.Equal(out.Blobs[1], []byte("beta")) ||
		out.OptionalKind == nil || *out.OptionalKind != Kind_KIND_B ||
		out.OptionalInner == nil || out.OptionalInner.B != "optional-local" {
		t.Fatalf("decoded fields wrong: %+v", out)
	}
	if h, ok := out.Auth.(*Shapes_Header); !ok || h.Header.GetValue() != "alice" {
		t.Fatalf("decoded auth oneof wrong: %#v", out.Auth)
	}
	if c, ok := out.ScalarChoice.(*Shapes_ChoiceText); !ok || c.ChoiceText != "scalar-oneof" {
		t.Fatalf("decoded scalar oneof wrong: %#v", out.ScalarChoice)
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
		!bytes.Equal(st.GetOptionalBlob(), []byte{192, 0, 2, 10}) ||
		len(st.Labels) != 2 || st.Labels["role"] != "proxy" ||
		st.OptionalI32 == nil || *st.OptionalI32 != -321 ||
		st.OptionalI64 == nil || *st.OptionalI64 != 51820 ||
		st.OptionalU32 == nil || *st.OptionalU32 != 8 ||
		st.OptionalDuration == nil || st.OptionalDuration.Seconds != 90 ||
		len(st.Resolved) != 2 || len(st.Resolved["peer-a"].Values) != 2 ||
		st.Resolved["peer-a"].Values[1] != "fd00::10" ||
		len(st.Kinds) != 2 || st.Kinds[1] != stock.Kind_KIND_B ||
		len(st.Blobs) != 2 || !bytes.Equal(st.Blobs[1], []byte("beta")) ||
		st.OptionalKind == nil || *st.OptionalKind != stock.Kind_KIND_B ||
		st.OptionalInner == nil || st.OptionalInner.B != "optional-local" {
		t.Fatalf("stock decoded our bytes wrong: %+v", st)
	}
	if h, ok := st.Auth.(*stock.Shapes_Header); !ok || h.Header.GetValue() != "alice" {
		t.Fatalf("stock decoded auth oneof wrong: %#v", st.Auth)
	}
	if c, ok := st.ScalarChoice.(*stock.Shapes_ChoiceText); !ok || c.ChoiceText != "scalar-oneof" {
		t.Fatalf("stock decoded scalar oneof wrong: %#v", st.ScalarChoice)
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
		!bytes.Equal(mine.GetOptionalBlob(), []byte{192, 0, 2, 10}) ||
		len(mine.Labels) != 2 || mine.Labels["role"] != "proxy" ||
		mine.OptionalI32 == nil || *mine.OptionalI32 != -321 ||
		mine.OptionalI64 == nil || *mine.OptionalI64 != 51820 ||
		mine.OptionalU32 == nil || *mine.OptionalU32 != 8 ||
		mine.OptionalDuration == nil || mine.OptionalDuration.Seconds != 90 ||
		len(mine.Resolved) != 2 || len(mine.Resolved["peer-a"].Values) != 2 ||
		mine.Resolved["peer-a"].Values[1] != "fd00::10" ||
		len(mine.Kinds) != 2 || mine.Kinds[1] != Kind_KIND_B ||
		len(mine.Blobs) != 2 || !bytes.Equal(mine.Blobs[1], []byte("beta")) ||
		mine.OptionalKind == nil || *mine.OptionalKind != Kind_KIND_B ||
		mine.OptionalInner == nil || mine.OptionalInner.B != "optional-local" {
		t.Fatalf("we decoded stock bytes wrong: %+v", mine)
	}
	if h, ok := mine.Auth.(*Shapes_Header); !ok || h.Header.GetValue() != "alice" {
		t.Fatalf("we decoded stock auth oneof wrong: %#v", mine.Auth)
	}
	if c, ok := mine.ScalarChoice.(*Shapes_ChoiceText); !ok || c.ChoiceText != "scalar-oneof" {
		t.Fatalf("we decoded stock scalar oneof wrong: %#v", mine.ScalarChoice)
	}
	if cm, ok := mine.Choice.(*Shapes_ChoiceMsg); !ok || cm.ChoiceMsg.B != "oneof" {
		t.Fatalf("we decoded stock oneof wrong: %#v", mine.Choice)
	}
}

func TestMessageOnlyOneofArms(t *testing.T) {
	for _, tt := range []struct {
		name string
		msg  *Shapes
		want func(*Shapes) bool
	}{
		{
			name: "password",
			msg:  &Shapes{Auth: &Shapes_Password{Password: &AuthPassword{Password: "secret"}}},
			want: func(got *Shapes) bool {
				v, ok := got.Auth.(*Shapes_Password)
				return ok && v.Password.GetPassword() == "secret"
			},
		},
		{
			name: "pin",
			msg:  &Shapes{Auth: &Shapes_Pin{Pin: &AuthPin{Pin: "123456"}}},
			want: func(got *Shapes) bool {
				v, ok := got.Auth.(*Shapes_Pin)
				return ok && v.Pin.GetPin() == "123456"
			},
		},
		{
			name: "header",
			msg:  &Shapes{Auth: &Shapes_Header{Header: &AuthHeader{Name: "x-user", Value: "alice"}}},
			want: func(got *Shapes) bool {
				v, ok := got.Auth.(*Shapes_Header)
				return ok && v.Header.GetName() == "x-user" && v.Header.GetValue() == "alice"
			},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			b, err := proto.Marshal(tt.msg)
			if err != nil {
				t.Fatal(err)
			}
			var got Shapes
			if err := proto.Unmarshal(b, &got); err != nil {
				t.Fatal(err)
			}
			if !tt.want(&got) {
				t.Fatalf("decoded auth oneof wrong: %#v", got.Auth)
			}

			var st stock.Shapes
			if err := proto.Unmarshal(b, &st); err != nil {
				t.Fatalf("stock unmarshal: %v", err)
			}
		})
	}
}

func TestScalarOneofArms(t *testing.T) {
	for _, tt := range []struct {
		name string
		msg  *Shapes
		want func(*Shapes) bool
	}{
		{
			name: "text",
			msg:  &Shapes{ScalarChoice: &Shapes_ChoiceText{ChoiceText: "hello"}},
			want: func(got *Shapes) bool {
				v, ok := got.ScalarChoice.(*Shapes_ChoiceText)
				return ok && v.ChoiceText == "hello"
			},
		},
		{
			name: "data",
			msg:  &Shapes{ScalarChoice: &Shapes_ChoiceData{ChoiceData: []byte{1, 2, 3}}},
			want: func(got *Shapes) bool {
				v, ok := got.ScalarChoice.(*Shapes_ChoiceData)
				return ok && bytes.Equal(v.ChoiceData, []byte{1, 2, 3})
			},
		},
		{
			name: "flag",
			msg:  &Shapes{ScalarChoice: &Shapes_ChoiceFlag{ChoiceFlag: true}},
			want: func(got *Shapes) bool {
				v, ok := got.ScalarChoice.(*Shapes_ChoiceFlag)
				return ok && v.ChoiceFlag
			},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			b, err := proto.Marshal(tt.msg)
			if err != nil {
				t.Fatal(err)
			}
			var got Shapes
			if err := proto.Unmarshal(b, &got); err != nil {
				t.Fatal(err)
			}
			if !tt.want(&got) {
				t.Fatalf("decoded scalar oneof wrong: %#v", got.ScalarChoice)
			}
		})
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
