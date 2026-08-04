package gen

import "fmt"

// ---------- Go type / literal helpers ----------

func scalarGoType(e Elem) string {
	switch e.Kind {
	case "bool", "int32", "int64", "uint32", "uint64", "float32", "float64", "string":
		return e.Kind
	case "enum":
		return e.Ref
	}
	return "any"
}

func elemGoType(e Elem) string {
	switch e.Kind {
	case "message":
		return "*" + e.Ref
	case "bytes":
		return "[]byte"
	}
	return scalarGoType(e)
}

func oneofFieldGoType(e Elem) string { return elemGoType(e) }

func fieldGoType(f Field) string {
	switch f.Card {
	case CardOptional:
		if f.Elem.Kind == "bytes" {
			return "[]byte"
		}
		if f.Elem.Kind == "message" {
			return elemGoType(f.Elem)
		}
		return "*" + scalarGoType(f.Elem)
	case CardRepeated:
		return "[]" + elemGoType(f.Elem)
	case CardMap:
		return "map[" + elemGoType(f.KeyElem) + "]" + elemGoType(f.ValElem)
	default:
		return elemGoType(f.Elem)
	}
}

func zeroLit(e Elem) string {
	switch e.Kind {
	case "bool":
		return "false"
	case "int32", "int64", "uint32", "uint64", "float32", "float64":
		return "0"
	case "string":
		return `""`
	case "enum":
		return e.Ref + "(0)"
	case "bytes", "message":
		return "nil"
	}
	return "nil"
}

// ---------- value-level codec fragments ----------

// scalarAppend returns `buf = protowire.Append...(buf, <expr(val)>)` (no tag).
func scalarAppend(buf, wire string, e Elem, val string) string {
	switch wire {
	case "varint":
		var arg string
		switch e.Kind {
		case "bool":
			arg = "embedBool(" + val + ")"
		case "int32", "enum":
			arg = "uint64(int64(" + val + "))"
		case "int64":
			arg = "uint64(" + val + ")"
		case "uint32":
			arg = "uint64(" + val + ")"
		case "uint64":
			arg = val
		}
		return fmt.Sprintf("%s = protowire.AppendVarint(%s, %s)", buf, buf, arg)
	case "zigzag32":
		return fmt.Sprintf("%s = protowire.AppendVarint(%s, protowire.EncodeZigZag(int64(%s)))", buf, buf, val)
	case "zigzag64":
		return fmt.Sprintf("%s = protowire.AppendVarint(%s, protowire.EncodeZigZag(%s))", buf, buf, val)
	case "fixed32":
		arg := val
		switch e.Kind {
		case "int32":
			arg = "uint32(" + val + ")"
		case "float32":
			arg = "math.Float32bits(" + val + ")"
		}
		return fmt.Sprintf("%s = protowire.AppendFixed32(%s, %s)", buf, buf, arg)
	case "fixed64":
		arg := val
		switch e.Kind {
		case "int64":
			arg = "uint64(" + val + ")"
		case "float64":
			arg = "math.Float64bits(" + val + ")"
		}
		return fmt.Sprintf("%s = protowire.AppendFixed64(%s, %s)", buf, buf, arg)
	case "bytes":
		if e.Kind == "string" {
			return fmt.Sprintf("%s = protowire.AppendString(%s, %s)", buf, buf, val)
		}
		return fmt.Sprintf("%s = protowire.AppendBytes(%s, %s)", buf, buf, val)
	}
	return "// UNSUPPORTED append " + wire
}

// scalarSize returns an int expression for the value size (no tag).
func scalarSize(wire string, e Elem, val string) string {
	switch wire {
	case "varint":
		switch e.Kind {
		case "bool":
			return "protowire.SizeVarint(embedBool(" + val + "))"
		case "int32", "enum":
			return "protowire.SizeVarint(uint64(int64(" + val + ")))"
		case "int64":
			return "protowire.SizeVarint(uint64(" + val + "))"
		case "uint32":
			return "protowire.SizeVarint(uint64(" + val + "))"
		case "uint64":
			return "protowire.SizeVarint(" + val + ")"
		}
	case "zigzag32":
		return "protowire.SizeVarint(protowire.EncodeZigZag(int64(" + val + ")))"
	case "zigzag64":
		return "protowire.SizeVarint(protowire.EncodeZigZag(" + val + "))"
	case "fixed32":
		return "protowire.SizeFixed32()"
	case "fixed64":
		return "protowire.SizeFixed64()"
	case "bytes":
		return "protowire.SizeBytes(len(" + val + "))"
	}
	return "0 /* UNSUPPORTED size */"
}

// consumeFunc returns the protowire.Consume* to read wire w into raw `v`.
func consumeFunc(wire string, e Elem) string {
	switch wire {
	case "varint", "zigzag32", "zigzag64":
		return "protowire.ConsumeVarint"
	case "fixed32":
		return "protowire.ConsumeFixed32"
	case "fixed64":
		return "protowire.ConsumeFixed64"
	case "bytes":
		if e.Kind == "string" {
			return "protowire.ConsumeString"
		}
		return "protowire.ConsumeBytes"
	}
	return "protowire.ConsumeBytes"
}

// decodeExpr converts raw var `v` into the Go value (scalars/enums only).
func decodeExpr(wire string, e Elem) string {
	switch wire {
	case "varint":
		switch e.Kind {
		case "bool":
			return "v != 0"
		case "int32":
			return "int32(v)"
		case "enum":
			return e.Ref + "(v)"
		case "int64":
			return "int64(v)"
		case "uint32":
			return "uint32(v)"
		case "uint64":
			return "v"
		}
	case "zigzag32":
		return "int32(protowire.DecodeZigZag(v))"
	case "zigzag64":
		return "protowire.DecodeZigZag(v)"
	case "fixed32":
		switch e.Kind {
		case "int32":
			return "int32(v)"
		case "float32":
			return "math.Float32frombits(v)"
		default:
			return "v"
		}
	case "fixed64":
		switch e.Kind {
		case "int64":
			return "int64(v)"
		case "float64":
			return "math.Float64frombits(v)"
		default:
			return "v"
		}
	case "bytes":
		return "v" // string
	}
	return "v"
}

// msgSizeExpr returns an int expression for a (nested) message value's size.
func msgSizeExpr(e Elem, val string) string {
	if e.External {
		return "extSize_" + mangle(e.Ref) + "(" + val + ")"
	}
	return val + ".sizeField()"
}

// emitSub writes `sub := <bytes>` for a nested message value.
func emitSub(x *w, indent string, e Elem, val string) {
	if e.External {
		x.p("%ssub := extAppend_%s(nil, %s)", indent, mangle(e.Ref), val)
	} else {
		x.p("%ssub, _ := %s.marshalAppend(nil)", indent, val)
	}
}

// emitMsgDecode writes statements decoding raw bytes into target (a *Msg lvalue).
// For use inside functions that return error (message unmarshalMsg).
func emitMsgDecode(x *w, indent string, e Elem, target, raw string) {
	if e.External {
		x.p("%s%s = extDecode_%s(%s)", indent, target, mangle(e.Ref), raw)
	} else {
		x.p("%smm := &%s{}; if err := mm.unmarshalMsg(%s); err != nil { return err }; %s = mm", indent, e.Ref, raw, target)
	}
}

// emitMsgDecodeNoErr is like emitMsgDecode but ignores the decode error, for use
// inside helpers that do not return error (map-entry / external decoders).
func emitMsgDecodeNoErr(x *w, indent string, e Elem, target, raw string) {
	if e.External {
		x.p("%s%s = extDecode_%s(%s)", indent, target, mangle(e.Ref), raw)
	} else {
		x.p("%smm := &%s{}; _ = mm.unmarshalMsg(%s); %s = mm", indent, e.Ref, raw, target)
	}
}

// zeroGuard returns a boolean expr that is true when `val` is non-default.
func zeroGuard(e Elem, val string) string {
	switch e.Kind {
	case "bool":
		return val
	case "string":
		return val + ` != ""`
	case "bytes":
		return "len(" + val + ") > 0"
	case "message":
		return val + " != nil"
	default: // numeric + enum
		return val + " != 0"
	}
}
