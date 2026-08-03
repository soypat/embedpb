package gen

// Reflection-free codec helpers for cross-package (well-known) messages such as
// timestamppb.Timestamp / durationpb.Duration. The stock type is kept (field
// types + getters unchanged), but its wire codec is emitted here operating on
// its public fields, so the protobuf runtime never reflects it. Supports the
// field shapes well-known types use: single scalars/bytes/messages, proto3
// optional scalars, and packed repeated numerics.
func renderExternals(x *w, exts []ExtMsg) {
	for _, e := range exts {
		renderExtSize(x, e)
		renderExtAppend(x, e)
		renderExtDecode(x, e)
		renderExtJSON(x, e)
	}
}

// renderExtJSON emits protojson-compatible JSON for a well-known type. Timestamp
// and Duration get their special string forms; any other external message falls
// back to a generic object over its public fields.
func renderExtJSON(x *w, e ExtMsg) {
	switch e.Ref {
	case "timestamppb.Timestamp":
		x.p("func extJSON_%s(dst []byte, v *%s) []byte {", e.Mangle, e.Ref)
		x.p("\tif v == nil { return append(dst, \"null\"...) }")
		x.p("\tt := time.Unix(v.Seconds, int64(v.Nanos)).UTC()")
		x.p("\tdst = append(dst, '\"')")
		x.p("\tdst = t.AppendFormat(dst, \"2006-01-02T15:04:05\")")
		x.p("\tif v.Nanos != 0 { dst = append(dst, '.'); dst = append(dst, jsonNanoFrac(v.Nanos)...) }")
		x.p("\tdst = append(dst, 'Z', '\"')")
		x.p("\treturn dst")
		x.p("}")
	case "durationpb.Duration":
		x.p("func extJSON_%s(dst []byte, v *%s) []byte {", e.Mangle, e.Ref)
		x.p("\tif v == nil { return append(dst, \"null\"...) }")
		x.p("\tsec, nanos := v.Seconds, v.Nanos")
		x.p("\tdst = append(dst, '\"')")
		x.p("\tif sec < 0 || nanos < 0 { dst = append(dst, '-'); if sec < 0 { sec = -sec }; if nanos < 0 { nanos = -nanos } }")
		x.p("\tdst = strconv.AppendInt(dst, sec, 10)")
		x.p("\tif nanos != 0 { dst = append(dst, '.'); dst = append(dst, jsonNanoFrac(nanos)...) }")
		x.p("\tdst = append(dst, 's', '\"')")
		x.p("\treturn dst")
		x.p("}")
	default:
		// generic object form over public fields.
		x.p("func extJSON_%s(dst []byte, v *%s) []byte {", e.Mangle, e.Ref)
		x.p("\tif v == nil { return append(dst, \"null\"...) }")
		x.p("\tdst = append(dst, '{')")
		for _, f := range e.Fields {
			x.p("\tdst = append(dst, %q...)", `"`+f.ProtoName+`":`)
			if f.Elem.Kind == "message" {
				emitMsgJSON(x, "\t", f.Elem, "v."+f.GoName)
			} else if f.Card == CardRepeated {
				x.p("\tdst = append(dst, '[')")
				x.p("\tfor i, e := range v.%s { if i > 0 { dst = append(dst, ',') }; _ = e", f.GoName)
				emitScalarJSON(x, "\t\t", f.Elem, "e")
				x.p("\t}")
				x.p("\tdst = append(dst, ']')")
			} else {
				emitScalarJSON(x, "\t", f.Elem, "v."+f.GoName)
			}
			x.p("\tdst = append(dst, ',')")
		}
		x.p("\tdst = jsonEndObj(dst)")
		x.p("\treturn dst")
		x.p("}")
	}
}

func renderExtSize(x *w, e ExtMsg) {
	x.p("func extSize_%s(v *%s) int {", e.Mangle, e.Ref)
	x.p("\tif v == nil { return 0 }")
	x.p("\tn := 0")
	for _, f := range e.Fields {
		val := "v." + f.GoName
		switch f.Card {
		case CardOptional:
			x.p("\tif %s != nil { n += protowire.SizeTag(%d) + %s }", val, f.Num, scalarSize(f.Wire, f.Elem, "*"+val))
		case CardRepeated:
			if isPackedNumeric(f) {
				x.p("\tif len(%s) > 0 { pn := 0; for _, v := range %s { pn += %s }; n += protowire.SizeTag(%d) + protowire.SizeBytes(pn) }", val, val, scalarSize(f.Wire, f.Elem, "v"), f.Num)
			}
		default:
			switch f.Elem.Kind {
			case "message":
				x.p("\tif %s != nil { n += protowire.SizeTag(%d) + protowire.SizeBytes(%s) }", val, f.Num, msgSizeExpr(f.Elem, val))
			case "bytes":
				x.p("\tif len(%s) > 0 { n += protowire.SizeTag(%d) + protowire.SizeBytes(len(%s)) }", val, f.Num, val)
			default:
				x.p("\tif %s { n += protowire.SizeTag(%d) + %s }", zeroGuard(f.Elem, val), f.Num, scalarSize(f.Wire, f.Elem, val))
			}
		}
	}
	x.p("\treturn n")
	x.p("}")
}

func renderExtAppend(x *w, e ExtMsg) {
	x.p("func extAppend_%s(b []byte, v *%s) []byte {", e.Mangle, e.Ref)
	x.p("\tif v == nil { return b }")
	for _, f := range e.Fields {
		val := "v." + f.GoName
		switch f.Card {
		case CardOptional:
			x.p("\tif %s != nil {", val)
			x.p("\t\tb = protowire.AppendTag(b, %d, %s)", f.Num, wireTypeConst(f.Wire))
			x.p("\t\t%s", scalarAppend("b", f.Wire, f.Elem, "*"+val))
			x.p("\t}")
		case CardRepeated:
			if isPackedNumeric(f) {
				x.p("\tif len(%s) > 0 {", val)
				x.p("\t\tvar packed []byte")
				x.p("\t\tfor _, v := range %s { %s }", val, scalarAppend("packed", f.Wire, f.Elem, "v"))
				x.p("\t\tb = protowire.AppendTag(b, %d, protowire.BytesType)", f.Num)
				x.p("\t\tb = protowire.AppendBytes(b, packed)")
				x.p("\t}")
			}
		default:
			switch f.Elem.Kind {
			case "message":
				x.p("\tif %s != nil {", val)
				emitSub(x, "\t\t", f.Elem, val)
				x.p("\t\tb = protowire.AppendTag(b, %d, protowire.BytesType)", f.Num)
				x.p("\t\tb = protowire.AppendBytes(b, sub)")
				x.p("\t}")
			case "bytes":
				x.p("\tif len(%s) > 0 { b = protowire.AppendTag(b, %d, protowire.BytesType); b = protowire.AppendBytes(b, %s) }", val, f.Num, val)
			default:
				x.p("\tif %s {", zeroGuard(f.Elem, val))
				x.p("\t\tb = protowire.AppendTag(b, %d, %s)", f.Num, wireTypeConst(f.Wire))
				x.p("\t\t%s", scalarAppend("b", f.Wire, f.Elem, val))
				x.p("\t}")
			}
		}
	}
	x.p("\treturn b")
	x.p("}")
}

func renderExtDecode(x *w, e ExtMsg) {
	x.p("func extDecode_%s(b []byte) *%s {", e.Mangle, e.Ref)
	x.p("\tm := &%s{}", e.Ref)
	x.p("\tfor len(b) > 0 {")
	x.p("\t\tnum, typ, n := protowire.ConsumeTag(b)")
	x.p("\t\tif n < 0 { break }")
	x.p("\t\tb = b[n:]")
	x.p("\t\tvar consumed int")
	x.p("\t\tswitch num {")
	for _, f := range e.Fields {
		val := "m." + f.GoName
		x.p("\t\tcase %d:", f.Num)
		switch f.Card {
		case CardOptional:
			x.p("\t\t\tv, k := %s(b); consumed = k; tmp := %s; %s = &tmp", consumeFunc(f.Wire, f.Elem), decodeExpr(f.Wire, f.Elem), val)
		case CardRepeated:
			if isPackedNumeric(f) {
				x.p("\t\t\tif typ == protowire.BytesType {")
				x.p("\t\t\t\tpk, k := protowire.ConsumeBytes(b); consumed = k")
				x.p("\t\t\t\tfor len(pk) > 0 { v, kk := %s(pk); if kk < 0 { break }; %s = append(%s, %s); pk = pk[kk:] }", consumeFunc(f.Wire, f.Elem), val, val, decodeExpr(f.Wire, f.Elem))
				x.p("\t\t\t} else { v, k := %s(b); consumed = k; %s = append(%s, %s) }", consumeFunc(f.Wire, f.Elem), val, val, decodeExpr(f.Wire, f.Elem))
			}
		default:
			switch f.Elem.Kind {
			case "message":
				x.p("\t\t\tv, k := protowire.ConsumeBytes(b); consumed = k")
				emitMsgDecodeNoErr(x, "\t\t\t", f.Elem, val, "v")
			case "bytes":
				x.p("\t\t\tv, k := protowire.ConsumeBytes(b); consumed = k; %s = append([]byte(nil), v...)", val)
			case "string":
				x.p("\t\t\tv, k := protowire.ConsumeString(b); consumed = k; %s = v", val)
			default:
				x.p("\t\t\tv, k := %s(b); consumed = k; %s = %s", consumeFunc(f.Wire, f.Elem), val, decodeExpr(f.Wire, f.Elem))
			}
		}
	}
	x.p("\t\tdefault:")
	x.p("\t\t\tconsumed = protowire.ConsumeFieldValue(num, typ, b)")
	x.p("\t\t}")
	x.p("\t\tif consumed < 0 { break }")
	x.p("\t\tb = b[consumed:]")
	x.p("\t}")
	x.p("\treturn m")
	x.p("}")
}
