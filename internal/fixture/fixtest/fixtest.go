// Package fixtest is the shared assertion harness for the generated fixtures.
//
// Every fixture asks the same questions and the answers do not depend on which
// proto the message came from, so they are stated once here.
package fixtest

import (
	"bytes"
	"encoding/json"
	"reflect"
	"testing"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

// Case is one sample message and the stock type to check it against.
type Case struct {
	Name string
	// Ours is an embedpb-generated message, populated.
	Ours proto.Message
	// Stock is an empty instance of its protoc-gen-go twin, e.g.
	// new(stock.LoginRequest). It is only ever used as a factory: the compared
	// stock value is decoded from Ours' bytes, so the two cannot drift.
	Stock proto.Message
}

// jsonAppender is the generated protojson-compatible encoder.
type jsonAppender interface {
	AppendJSON([]byte) ([]byte, error)
}

// protojsonOpts are the options netbird's wasm bridge uses.
var protojsonOpts = protojson.MarshalOptions{
	EmitUnpopulated: true,
	UseProtoNames:   true,
	AllowPartial:    true,
}

// Run asserts, for each case: proto.Size agrees with the marshal; our bytes
// decode and re-encode to themselves; the message survives a trip through the
// stock runtime unchanged; and AppendJSON agrees with protojson.
func Run(t *testing.T, cases []Case) {
	t.Helper()
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			ours := mustMarshal(t, "ours", c.Ours)
			if n := proto.Size(c.Ours); n != len(ours) {
				t.Errorf("proto.Size = %d, but marshal wrote %d bytes", n, len(ours))
			}

			back := fresh(t, c.Ours)
			if err := proto.Unmarshal(ours, back); err != nil {
				t.Fatalf("our unmarshal of our bytes: %v", err)
			}
			if got := mustMarshal(t, "our re-marshal", back); !bytes.Equal(got, ours) {
				t.Fatalf("our decode/encode is not a round trip\n got  %x\n want %x", got, ours)
			}

			// The stock runtime stands in for every peer on the other end. Its
			// bytes are not compared to ours directly: wire order is free and
			// stock appends oneofs after other fields. What must hold is that
			// the message survives the trip.
			st := fresh(t, c.Stock)
			if err := proto.Unmarshal(ours, st); err != nil {
				t.Fatalf("stock unmarshal of our bytes: %v", err)
			}
			mine := fresh(t, c.Ours)
			if err := proto.Unmarshal(mustMarshal(t, "stock re-marshal", st), mine); err != nil {
				t.Fatalf("our unmarshal of stock bytes: %v", err)
			}
			if got := mustMarshal(t, "round trip", mine); !bytes.Equal(got, ours) {
				t.Fatalf("stock round trip changed the message\n got  %x\n want %x", got, ours)
			}

			checkJSON(t, c.Ours, st)
		})
	}
}

// checkJSON compares our JSON against protojson's, as decoded values:
// protojson injects a random space (internal/detrand), so its bytes are
// deliberately unstable.
func checkJSON(t *testing.T, ours, stock proto.Message) {
	t.Helper()
	ja, ok := ours.(jsonAppender)
	if !ok {
		t.Fatalf("%T has no AppendJSON", ours)
	}
	got, err := ja.AppendJSON(nil)
	if err != nil {
		t.Fatalf("AppendJSON: %v", err)
	}
	want, err := protojsonOpts.Marshal(stock)
	if err != nil {
		t.Fatalf("protojson.Marshal: %v", err)
	}
	var gotV, wantV any
	if err := json.Unmarshal(got, &gotV); err != nil {
		t.Fatalf("our JSON is invalid: %v\n%s", err, got)
	}
	if err := json.Unmarshal(want, &wantV); err != nil {
		t.Fatalf("protojson output is invalid: %v", err)
	}
	if !reflect.DeepEqual(gotV, wantV) {
		t.Errorf("JSON mismatch\n got  %s\n want %s", got, want)
	}
}

// fresh returns an empty instance of m's type. Both codecs support it: the
// generated shim implements New(), so the table needs no factory funcs.
func fresh(t *testing.T, m proto.Message) proto.Message {
	t.Helper()
	if m == nil {
		t.Fatal("nil message in fixtest.Case")
	}
	return m.ProtoReflect().New().Interface()
}

func mustMarshal(t *testing.T, what string, m proto.Message) []byte {
	t.Helper()
	b, err := proto.Marshal(m)
	if err != nil {
		t.Fatalf("marshal %s: %v", what, err)
	}
	return b
}
