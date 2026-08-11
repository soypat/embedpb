package gen

import (
	"fmt"
	"strings"
)

func entryHelperName(msg, field string) string { return fmt.Sprintf("mapEntry_%s_%s", msg, field) }

// ---------- Size ----------

func renderSize(x *w, m Message) {
	x.p("func (m *%s) sizeField() int {", m.GoName)
	x.p("\tn := 0")
	for _, f := range m.Fields {
		if f.Oneof != nil {
			x.p("\tswitch c := m.%s.(type) {", f.Oneof.GoName)
			for _, c := range f.Oneof.Cases {
				x.p("\tcase *%s:", c.WrapperGo)
				x.p("\t\t_ = c")
				if c.Elem.Kind == "message" {
					x.p("\t\tn += protowire.SizeTag(%d) + protowire.SizeBytes(%s)", c.Num, msgSizeExpr(c.Elem, "c."+c.FieldGo))
				} else {
					x.p("\t\tn += protowire.SizeTag(%d) + %s", c.Num, scalarSize(c.Wire, c.Elem, "c."+c.FieldGo))
				}
			}
			x.p("\t}")
			continue
		}
		val := "m." + f.GoName
		switch f.Card {
		case CardMap:
			x.p("\tfor k, v := range %s {", val)
			x.p("\t\tn += protowire.SizeTag(%d) + protowire.SizeBytes(%s(k, v))", f.Num, entryHelperName(m.GoName, f.GoName))
			x.p("\t}")
		case CardRepeated:
			if isPackedNumeric(f) {
				x.p("\tif len(%s) > 0 {", val)
				x.p("\t\tpn := 0")
				x.p("\t\tfor _, v := range %s { pn += %s }", val, scalarSize(f.Wire, f.Elem, "v"))
				x.p("\t\tn += protowire.SizeTag(%d) + protowire.SizeBytes(pn)", f.Num)
				x.p("\t}")
			} else if f.Elem.Kind == "message" {
				x.p("\tfor _, v := range %s { n += protowire.SizeTag(%d) + protowire.SizeBytes(%s) }", val, f.Num, msgSizeExpr(f.Elem, "v"))
			} else {
				x.p("\tfor _, v := range %s { n += protowire.SizeTag(%d) + %s }", val, f.Num, scalarSize(f.Wire, f.Elem, "v"))
			}
		case CardOptional:
			x.p("\tif %s != nil {", val)
			if f.Elem.Kind == "message" {
				x.p("\t\tn += protowire.SizeTag(%d) + protowire.SizeBytes(%s)", f.Num, msgSizeExpr(f.Elem, val))
			} else if f.Elem.Kind == "bytes" {
				x.p("\t\tn += protowire.SizeTag(%d) + protowire.SizeBytes(len(%s))", f.Num, val)
			} else {
				x.p("\t\tn += protowire.SizeTag(%d) + %s", f.Num, scalarSize(f.Wire, f.Elem, "*"+val))
			}
			x.p("\t}")
		default:
			if f.Elem.Kind == "message" {
				x.p("\tif %s != nil { n += protowire.SizeTag(%d) + protowire.SizeBytes(%s) }", val, f.Num, msgSizeExpr(f.Elem, val))
			} else {
				x.p("\tif %s { n += protowire.SizeTag(%d) + %s }", zeroGuard(f.Elem, val), f.Num, scalarSize(f.Wire, f.Elem, val))
			}
		}
	}
	x.p("\tn += len(m.unknownFields)")
	x.p("\treturn n")
	x.p("}")
	x.p("")
}

// ---------- Marshal ----------

func renderMarshal(x *w, m Message) {
	x.p("func (m *%s) marshalAppend(b []byte) ([]byte, error) {", m.GoName)
	for _, f := range m.Fields {
		if f.Oneof != nil {
			x.p("\tswitch c := m.%s.(type) {", f.Oneof.GoName)
			for _, c := range f.Oneof.Cases {
				x.p("\tcase *%s:", c.WrapperGo)
				if c.Elem.Kind == "message" {
					emitSub(x, "\t\t", c.Elem, "c."+c.FieldGo)
					x.p("\t\tb = protowire.AppendTag(b, %d, protowire.BytesType)", c.Num)
					x.p("\t\tb = protowire.AppendBytes(b, sub)")
				} else {
					x.p("\t\tb = protowire.AppendTag(b, %d, %s)", c.Num, wireTypeConst(c.Wire))
					x.p("\t\t%s", scalarAppend("b", c.Wire, c.Elem, "c."+c.FieldGo))
				}
			}
			x.p("\t}")
			continue
		}
		val := "m." + f.GoName
		switch f.Card {
		case CardMap:
			renderMapMarshal(x, m, f)
		case CardRepeated:
			if isPackedNumeric(f) {
				x.p("\tif len(%s) > 0 {", val)
				x.p("\t\tvar packed []byte")
				x.p("\t\tfor _, v := range %s { %s }", val, scalarAppend("packed", f.Wire, f.Elem, "v"))
				x.p("\t\tb = protowire.AppendTag(b, %d, protowire.BytesType)", f.Num)
				x.p("\t\tb = protowire.AppendBytes(b, packed)")
				x.p("\t}")
			} else if f.Elem.Kind == "message" {
				x.p("\tfor _, v := range %s {", val)
				emitSub(x, "\t\t", f.Elem, "v")
				x.p("\t\tb = protowire.AppendTag(b, %d, protowire.BytesType)", f.Num)
				x.p("\t\tb = protowire.AppendBytes(b, sub)")
				x.p("\t}")
			} else {
				x.p("\tfor _, v := range %s {", val)
				x.p("\t\tb = protowire.AppendTag(b, %d, %s)", f.Num, wireTypeConst(f.Wire))
				x.p("\t\t%s", scalarAppend("b", f.Wire, f.Elem, "v"))
				x.p("\t}")
			}
		case CardOptional:
			x.p("\tif %s != nil {", val)
			if f.Elem.Kind == "message" {
				emitSub(x, "\t\t", f.Elem, val)
				x.p("\t\tb = protowire.AppendTag(b, %d, protowire.BytesType)", f.Num)
				x.p("\t\tb = protowire.AppendBytes(b, sub)")
			} else if f.Elem.Kind == "bytes" {
				x.p("\t\tb = protowire.AppendTag(b, %d, %s)", f.Num, wireTypeConst(f.Wire))
				x.p("\t\tb = protowire.AppendBytes(b, %s)", val)
			} else {
				x.p("\t\tb = protowire.AppendTag(b, %d, %s)", f.Num, wireTypeConst(f.Wire))
				x.p("\t\t%s", scalarAppend("b", f.Wire, f.Elem, "*"+val))
			}
			x.p("\t}")
		default:
			if f.Elem.Kind == "message" {
				x.p("\tif %s != nil {", val)
				emitSub(x, "\t\t", f.Elem, val)
				x.p("\t\tb = protowire.AppendTag(b, %d, protowire.BytesType)", f.Num)
				x.p("\t\tb = protowire.AppendBytes(b, sub)")
				x.p("\t}")
			} else {
				x.p("\tif %s {", zeroGuard(f.Elem, val))
				x.p("\t\tb = protowire.AppendTag(b, %d, %s)", f.Num, wireTypeConst(f.Wire))
				x.p("\t\t%s", scalarAppend("b", f.Wire, f.Elem, val))
				x.p("\t}")
			}
		}
	}
	x.p("\tb = append(b, m.unknownFields...)")
	x.p("\treturn b, nil")
	x.p("}")
	x.p("")
}

func renderMapMarshal(x *w, m Message, f Field) {
	val := "m." + f.GoName
	kt := elemGoType(f.KeyElem)
	x.p("\tif len(%s) > 0 {", val)
	x.p("\t\tkeys := make([]%s, 0, len(%s))", kt, val)
	x.p("\t\tfor k := range %s { keys = append(keys, k) }", val)
	x.p("\t\tsort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })")
	x.p("\t\tfor _, k := range keys {")
	x.p("\t\t\tb = protowire.AppendTag(b, %d, protowire.BytesType)", f.Num)
	x.p("\t\t\tb = protowire.AppendBytes(b, %s(nil, k, %s[k]))", appendEntryName(m.GoName, f.GoName), val)
	x.p("\t\t}")
	x.p("\t}")
}

// ---------- Unmarshal ----------

func renderUnmarshal(x *w, m Message) {
	x.p("func (m *%s) unmarshalMsg(b []byte) error {", m.GoName)
	x.p("\tfor len(b) > 0 {")
	x.p("\t\tnum, typ, n := protowire.ConsumeTag(b)")
	x.p("\t\tif n < 0 { return protowire.ParseError(n) }")
	x.p("\t\tstart := b")
	x.p("\t\tb = b[n:]")
	x.p("\t\tvar consumed int")
	x.p("\t\tswitch num {")
	for _, f := range m.Fields {
		if f.Oneof != nil {
			for _, c := range f.Oneof.Cases {
				x.p("\t\tcase %d:", c.Num)
				if c.Elem.Kind == "message" {
					x.p("\t\t\tv, k := protowire.ConsumeBytes(b); consumed = k")
					x.p("\t\t\tif consumed >= 0 {")
					x.p("\t\t\t\tvar mv %s", oneofFieldGoType(c.Elem))
					emitMsgDecode(x, "\t\t\t\t", c.Elem, "mv", "v")
					x.p("\t\t\t\tm.%s = &%s{%s: mv}", f.Oneof.GoName, c.WrapperGo, c.FieldGo)
					x.p("\t\t\t}")
				} else {
					x.p("\t\t\tv, k := %s(b); consumed = k", consumeFunc(c.Wire, c.Elem))
					x.p("\t\t\tm.%s = &%s{%s: %s}", f.Oneof.GoName, c.WrapperGo, c.FieldGo, castTo(c.Elem, decodeExpr(c.Wire, c.Elem)))
				}
			}
			continue
		}
		x.p("\t\tcase %d:", f.Num)
		renderFieldUnmarshal(x, m, f)
	}
	x.p("\t\tdefault:")
	x.p("\t\t\tskip := protowire.ConsumeFieldValue(num, typ, b)")
	x.p("\t\t\tif skip < 0 { return protowire.ParseError(skip) }")
	x.p("\t\t\tm.unknownFields = append(m.unknownFields, start[:n+skip]...)")
	x.p("\t\t\tb = b[skip:]")
	x.p("\t\t\tcontinue")
	x.p("\t\t}")
	x.p("\t\tif consumed < 0 { return protowire.ParseError(consumed) }")
	x.p("\t\tb = b[consumed:]")
	x.p("\t}")
	x.p("\treturn nil")
	x.p("}")
	x.p("")

	// map entry helpers
	for _, f := range m.Fields {
		if f.Card == CardMap {
			renderMapEntryHelpers(x, m, f)
		}
	}
}

func renderFieldUnmarshal(x *w, m Message, f Field) {
	val := "m." + f.GoName
	switch f.Card {
	case CardMap:
		x.p("\t\t\tv, k := protowire.ConsumeBytes(b); consumed = k")
		x.p("\t\t\tif consumed >= 0 {")
		x.p("\t\t\t\tif %s == nil { %s = make(%s) }", val, val, fieldGoType(f))
		x.p("\t\t\t\tmk, mv := %s(v); %s[mk] = mv", decodeEntryName(m.GoName, f.GoName), val)
		x.p("\t\t\t}")
	case CardRepeated:
		if isPackedNumeric(f) {
			x.p("\t\t\tif typ == protowire.BytesType {")
			x.p("\t\t\t\tpk, k := protowire.ConsumeBytes(b); consumed = k")
			x.p("\t\t\t\tfor len(pk) > 0 { v, kk := %s(pk); if kk < 0 { return protowire.ParseError(kk) }; %s = append(%s, %s); pk = pk[kk:] }", consumeFunc(f.Wire, f.Elem), val, val, castTo(f.Elem, decodeExpr(f.Wire, f.Elem)))
			x.p("\t\t\t} else {")
			x.p("\t\t\t\tv, k := %s(b); consumed = k; %s = append(%s, %s)", consumeFunc(f.Wire, f.Elem), val, val, castTo(f.Elem, decodeExpr(f.Wire, f.Elem)))
			x.p("\t\t\t}")
		} else if f.Elem.Kind == "message" {
			x.p("\t\t\tv, k := protowire.ConsumeBytes(b); consumed = k")
			x.p("\t\t\tif consumed >= 0 {")
			x.p("\t\t\t\tvar mv %s", elemGoType(f.Elem))
			emitMsgDecode(x, "\t\t\t\t", f.Elem, "mv", "v")
			x.p("\t\t\t\t%s = append(%s, mv)", val, val)
			x.p("\t\t\t}")
		} else if f.Elem.Kind == "bytes" {
			x.p("\t\t\tv, k := protowire.ConsumeBytes(b); consumed = k")
			x.p("\t\t\t%s = append(%s, append([]byte(nil), v...))", val, val)
		} else {
			x.p("\t\t\tv, k := %s(b); consumed = k; %s = append(%s, %s)", consumeFunc(f.Wire, f.Elem), val, val, castTo(f.Elem, decodeExpr(f.Wire, f.Elem)))
		}
	case CardOptional:
		if f.Elem.Kind == "message" {
			x.p("\t\t\tv, k := protowire.ConsumeBytes(b); consumed = k")
			x.p("\t\t\tif consumed >= 0 {")
			emitMsgDecode(x, "\t\t\t\t", f.Elem, val, "v")
			x.p("\t\t\t}")
		} else if f.Elem.Kind == "bytes" {
			// []byte{}, not []byte(nil): nil is the unset representation here,
			// so a present-but-empty value must stay non-nil.
			x.p("\t\t\tv, k := protowire.ConsumeBytes(b); consumed = k")
			x.p("\t\t\t%s = append([]byte{}, v...)", val)
		} else {
			x.p("\t\t\tv, k := %s(b); consumed = k", consumeFunc(f.Wire, f.Elem))
			x.p("\t\t\ttmp := %s; %s = &tmp", castTo(f.Elem, decodeExpr(f.Wire, f.Elem)), val)
		}
	default:
		switch f.Elem.Kind {
		case "message":
			x.p("\t\t\tv, k := protowire.ConsumeBytes(b); consumed = k")
			x.p("\t\t\tif consumed >= 0 {")
			emitMsgDecode(x, "\t\t\t\t", f.Elem, val, "v")
			x.p("\t\t\t}")
		case "bytes":
			x.p("\t\t\tv, k := protowire.ConsumeBytes(b); consumed = k")
			x.p("\t\t\t%s = append([]byte(nil), v...)", val)
		case "string":
			x.p("\t\t\tv, k := protowire.ConsumeString(b); consumed = k; %s = v", val)
		default:
			x.p("\t\t\tv, k := %s(b); consumed = k; %s = %s", consumeFunc(f.Wire, f.Elem), val, castTo(f.Elem, decodeExpr(f.Wire, f.Elem)))
		}
	}
}

// castTo wraps decodeExpr result — decodeExpr already casts, so identity.
func castTo(_ Elem, expr string) string { return expr }

// ---------- map entry helpers ----------

func appendEntryName(msg, field string) string { return fmt.Sprintf("appendEntry_%s_%s", msg, field) }
func decodeEntryName(msg, field string) string { return fmt.Sprintf("decodeEntry_%s_%s", msg, field) }

func renderMapEntryHelpers(x *w, m Message, f Field) {
	kt := elemGoType(f.KeyElem)
	vt := elemGoType(f.ValElem)

	// size
	x.p("func %s(k %s, v %s) int {", entryHelperName(m.GoName, f.GoName), kt, vt)
	x.p("\tn := 0")
	x.p("\tif %s { n += protowire.SizeTag(1) + %s }", zeroGuard(f.KeyElem, "k"), keyValSize(f.KeyWire, f.KeyElem, "k"))
	if f.ValElem.Kind == "message" {
		x.p("\tif v != nil { n += protowire.SizeTag(2) + protowire.SizeBytes(%s) }", msgSizeExpr(f.ValElem, "v"))
	} else {
		x.p("\tif %s { n += protowire.SizeTag(2) + %s }", zeroGuard(f.ValElem, "v"), keyValSize(f.ValWire, f.ValElem, "v"))
	}
	x.p("\treturn n")
	x.p("}")

	// append
	x.p("func %s(b []byte, k %s, v %s) []byte {", appendEntryName(m.GoName, f.GoName), kt, vt)
	x.p("\tif %s {", zeroGuard(f.KeyElem, "k"))
	x.p("\t\tb = protowire.AppendTag(b, 1, %s)", wireTypeConst(f.KeyWire))
	x.p("\t\t%s", scalarAppend("b", f.KeyWire, f.KeyElem, "k"))
	x.p("\t}")
	if f.ValElem.Kind == "message" {
		x.p("\tif v != nil {")
		emitSub(x, "\t\t", f.ValElem, "v")
		x.p("\t\tb = protowire.AppendTag(b, 2, protowire.BytesType)")
		x.p("\t\tb = protowire.AppendBytes(b, sub)")
		x.p("\t}")
	} else {
		x.p("\tif %s {", zeroGuard(f.ValElem, "v"))
		x.p("\t\tb = protowire.AppendTag(b, 2, %s)", wireTypeConst(f.ValWire))
		x.p("\t\t%s", scalarAppend("b", f.ValWire, f.ValElem, "v"))
		x.p("\t}")
	}
	x.p("\treturn b")
	x.p("}")

	// decode
	x.p("func %s(b []byte) (%s, %s) {", decodeEntryName(m.GoName, f.GoName), kt, vt)
	x.p("\tvar k %s", kt)
	x.p("\tvar v %s", vt)
	x.p("\tfor len(b) > 0 {")
	x.p("\t\tnum, typ, n := protowire.ConsumeTag(b)")
	x.p("\t\tif n < 0 { break }")
	x.p("\t\tb = b[n:]")
	x.p("\t\tvar consumed int")
	x.p("\t\tswitch num {")
	x.p("\t\tcase 1:")
	if f.KeyElem.Kind == "string" {
		x.p("\t\t\tvv, kk := protowire.ConsumeString(b); consumed = kk; k = vv")
	} else {
		x.p("\t\t\tvv, kk := %s(b); consumed = kk; k = %s", consumeFunc(f.KeyWire, f.KeyElem), keyValDecode(f.KeyWire, f.KeyElem, "vv"))
	}
	x.p("\t\tcase 2:")
	if f.ValElem.Kind == "message" {
		x.p("\t\t\tvv, kk := protowire.ConsumeBytes(b); consumed = kk")
		emitMsgDecodeNoErr(x, "\t\t\t", f.ValElem, "v", "vv")
	} else if f.ValElem.Kind == "bytes" {
		x.p("\t\t\tvv, kk := protowire.ConsumeBytes(b); consumed = kk; v = append([]byte(nil), vv...)")
	} else if f.ValElem.Kind == "string" {
		x.p("\t\t\tvv, kk := protowire.ConsumeString(b); consumed = kk; v = vv")
	} else {
		x.p("\t\t\tvv, kk := %s(b); consumed = kk; v = %s", consumeFunc(f.ValWire, f.ValElem), keyValDecode(f.ValWire, f.ValElem, "vv"))
	}
	x.p("\t\tdefault:")
	x.p("\t\t\tconsumed = protowire.ConsumeFieldValue(num, typ, b)")
	x.p("\t\t}")
	x.p("\t\tif consumed < 0 { break }")
	x.p("\t\tb = b[consumed:]")
	x.p("\t}")
	if f.ValElem.Kind == "message" {
		x.p("\tif v == nil { v = &%s{} }", f.ValElem.Ref)
	}
	x.p("\treturn k, v")
	x.p("}")
	x.p("")
}

// keyValSize/Decode mirror scalarSize/decodeExpr but the raw var for decode is
// named `vv` (to avoid clash with the `v` value var).
func keyValSize(wire string, e Elem, val string) string { return scalarSize(wire, e, val) }

func keyValDecode(wire string, e Elem, raw string) string {
	// decodeExpr assumes raw var name "v"; substitute.
	expr := decodeExpr(wire, e)
	return replaceVar(expr, raw)
}

func replaceVar(expr, raw string) string {
	// crude: decodeExpr only uses bare "v" for the raw; replace whole-word v.
	var out strings.Builder
	for i := 0; i < len(expr); i++ {
		if expr[i] == 'v' && !isIdentChar(prevByte(expr, i)) && !isIdentChar(nextByte(expr, i)) {
			out.WriteString(raw)
			continue
		}
		out.WriteByte(expr[i])
	}
	return out.String()
}

func prevByte(s string, i int) byte {
	if i == 0 {
		return ' '
	}
	return s[i-1]
}
func nextByte(s string, i int) byte {
	if i+1 >= len(s) {
		return ' '
	}
	return s[i+1]
}
func isIdentChar(b byte) bool {
	return b == '_' || (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z') || (b >= '0' && b <= '9')
}

func isPackedNumeric(f Field) bool {
	switch f.Elem.Kind {
	case "bool", "int32", "int64", "uint32", "uint64", "float32", "float64", "enum":
		return f.Card == CardRepeated
	}
	return false
}
