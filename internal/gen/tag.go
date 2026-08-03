package gen

import (
	"strconv"
	"strings"
)

// protoTag holds the parsed pieces of a `protobuf:"..."` struct tag.
type protoTag struct {
	Wire   string // varint zigzag32 zigzag64 fixed32 fixed64 bytes group
	Num    int
	Label  string // opt rep req
	Packed bool
	Oneof  bool
	Enum   string // fully-qualified enum name from enum=... (may be empty)
	Name   string // proto field name from name=... (for protojson UseProtoNames)
}

// parseProtoTag parses the value of a `protobuf` struct tag.
// Example: "bytes,2,opt,name=key,proto3" or "varint,16,rep,packed,...".
func parseProtoTag(v string) (protoTag, bool) {
	if v == "" {
		return protoTag{}, false
	}
	parts := strings.Split(v, ",")
	if len(parts) < 2 {
		return protoTag{}, false
	}
	t := protoTag{Wire: parts[0]}
	n, err := strconv.Atoi(parts[1])
	if err != nil {
		return protoTag{}, false
	}
	t.Num = n
	for _, p := range parts[2:] {
		switch {
		case p == "opt" || p == "rep" || p == "req":
			t.Label = p
		case p == "packed":
			t.Packed = true
		case p == "oneof":
			t.Oneof = true
		case strings.HasPrefix(p, "enum="):
			t.Enum = strings.TrimPrefix(p, "enum=")
		case strings.HasPrefix(p, "name="):
			t.Name = strings.TrimPrefix(p, "name=")
		}
	}
	return t, true
}

// wireTypeConst maps a proto wire kind to the protowire type constant name.
func wireTypeConst(wire string) string {
	switch wire {
	case "varint", "zigzag32", "zigzag64":
		return "protowire.VarintType"
	case "fixed32":
		return "protowire.Fixed32Type"
	case "fixed64":
		return "protowire.Fixed64Type"
	case "bytes":
		return "protowire.BytesType"
	}
	return "protowire.BytesType"
}
