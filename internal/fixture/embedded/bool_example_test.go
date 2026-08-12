package shapespb

import (
	"bytes"
	"testing"

	stock "github.com/soypat/embedpb/internal/fixture/stock"

	"google.golang.org/protobuf/proto"
)

// TestBoolFieldExample is intentionally tiny: it shows the simplest possible
// flow for an embedpb-generated message.
func TestBoolFieldExample(t *testing.T) {
	msg := &Shapes{Flag: true}

	got, err := proto.Marshal(msg)
	if err != nil {
		t.Fatal(err)
	}

	// Shapes.flag is field 11. In protobuf wire format:
	//   0x58 = field 11, varint
	//   0x01 = true
	want := []byte{0x58, 0x01}

	if !bytes.Equal(got, want) {
		t.Fatalf("wire mismatch: got %x, want %x", got, want)
	}

	var decoded Shapes
	if err := proto.Unmarshal(got, &decoded); err != nil {
		t.Fatal(err)
	}
	if !decoded.Flag {
		t.Fatalf("decoded Flag = false, want true")
	}

	var stockDecoded stock.Shapes
	if err := proto.Unmarshal(got, &stockDecoded); err != nil {
		t.Fatal(err)
	}
	if !stockDecoded.Flag {
		t.Fatalf("stock decoded Flag = false, want true")
	}
}
