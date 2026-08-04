package gen

// protojson-compatible JSON emission (UseProtoNames + EmitUnpopulated), the
// options netbird's wasm bridge uses at client/wasm/cmd/main.go. Emits, per
// message, MarshalJSON + AppendJSON(dst) so the JS side gets the same shape the
// stock protojson produced, with zero reflect. Rules matched: proto field
// names, 64-bit integers as quoted strings, bytes as std base64, enums as
// value-name strings, well-known Timestamp/Duration as their special string
// forms, all fields always emitted (EmitUnpopulated).

func renderJSON(x *w, m Message) {
	x.p("func (m *%s) MarshalJSON() ([]byte, error) { return m.AppendJSON(nil) }", m.GoName)
	x.p("func (m *%s) AppendJSON(dst []byte) ([]byte, error) {", m.GoName)
	x.p("\tvar err error")
	x.p("\t_ = err")
	x.p("\tdst = append(dst, '{')")
	for _, f := range m.Fields {
		emitFieldJSON(x, m, f)
	}
	x.p("\tdst = jsonEndObj(dst)")
	x.p("\treturn dst, nil")
	x.p("}")
	x.p("")
}

func emitFieldJSON(x *w, m Message, f Field) {
	if f.Oneof != nil {
		x.p("\tswitch c := m.%s.(type) {", f.Oneof.GoName)
		for _, c := range f.Oneof.Cases {
			x.p("\tcase *%s:", c.WrapperGo)
			x.p("\t\tdst = append(dst, %q...)", `"`+c.ProtoName+`":`)
			if c.Elem.Kind == "message" {
				emitMsgJSON(x, "\t\t", c.Elem, "c."+c.FieldGo)
			} else {
				emitScalarJSON(x, "\t\t", c.Elem, "c."+c.FieldGo)
			}
			x.p("\t\tdst = append(dst, ',')")
		}
		x.p("\t}")
		return
	}

	val := "m." + f.GoName
	if f.Card == CardOptional {
		x.p("\tif %s != nil {", val)
		x.p("\t\tdst = append(dst, %q...)", `"`+f.ProtoName+`":`)
		if f.Elem.Kind == "bytes" {
			emitScalarJSON(x, "\t\t", f.Elem, val)
		} else {
			emitScalarJSON(x, "\t\t", f.Elem, "*"+val)
		}
		x.p("\t\tdst = append(dst, ',')")
		x.p("\t}")
		return
	}

	x.p("\tdst = append(dst, %q...)", `"`+f.ProtoName+`":`)
	switch f.Card {
	case CardMap:
		x.p("\tdst = append(dst, '{')")
		x.p("\t{")
		x.p("\t\tmk := make([]%s, 0, len(%s))", elemGoType(f.KeyElem), val)
		x.p("\t\tfor k := range %s { mk = append(mk, k) }", val)
		x.p("\t\tsort.Slice(mk, func(i, j int) bool { return mk[i] < mk[j] })")
		x.p("\t\tfor i, k := range mk {")
		x.p("\t\t\tif i > 0 { dst = append(dst, ',') }")
		emitMapKeyJSON(x, "\t\t\t", f.KeyElem, "k")
		x.p("\t\t\tdst = append(dst, ':')")
		x.p("\t\t\tv := %s[k]", val)
		if f.ValElem.Kind == "message" {
			emitMsgJSON(x, "\t\t\t", f.ValElem, "v")
		} else {
			emitScalarJSON(x, "\t\t\t", f.ValElem, "v")
		}
		x.p("\t\t}")
		x.p("\t}")
		x.p("\tdst = append(dst, '}')")
	case CardRepeated:
		x.p("\tdst = append(dst, '[')")
		x.p("\tfor i, v := range %s {", val)
		x.p("\t\tif i > 0 { dst = append(dst, ',') }")
		if f.Elem.Kind == "message" {
			emitMsgJSON(x, "\t\t", f.Elem, "v")
		} else {
			emitScalarJSON(x, "\t\t", f.Elem, "v")
		}
		x.p("\t}")
		x.p("\tdst = append(dst, ']')")
	default:
		if f.Elem.Kind == "message" {
			emitMsgJSON(x, "\t", f.Elem, val)
		} else {
			emitScalarJSON(x, "\t", f.Elem, val)
		}
	}
	x.p("\tdst = append(dst, ',')")
}

// emitScalarJSON writes JSON for a scalar/bytes/enum value expression.
func emitScalarJSON(x *w, ind string, e Elem, val string) {
	switch e.Kind {
	case "bool":
		x.p("%sdst = strconv.AppendBool(dst, %s)", ind, val)
	case "int32":
		x.p("%sdst = strconv.AppendInt(dst, int64(%s), 10)", ind, val)
	case "uint32":
		x.p("%sdst = strconv.AppendUint(dst, uint64(%s), 10)", ind, val)
	case "int64":
		x.p("%sdst = append(dst, '\"'); dst = strconv.AppendInt(dst, int64(%s), 10); dst = append(dst, '\"')", ind, val)
	case "uint64":
		x.p("%sdst = append(dst, '\"'); dst = strconv.AppendUint(dst, uint64(%s), 10); dst = append(dst, '\"')", ind, val)
	case "float32":
		x.p("%sdst = strconv.AppendFloat(dst, float64(%s), 'g', -1, 32)", ind, val)
	case "float64":
		x.p("%sdst = strconv.AppendFloat(dst, %s, 'g', -1, 64)", ind, val)
	case "string":
		x.p("%sdst = jsonString(dst, %s)", ind, val)
	case "bytes":
		x.p("%sdst = jsonBytes(dst, %s)", ind, val)
	case "enum":
		x.p("%sif s, ok := %s_name[int32(%s)]; ok { dst = jsonString(dst, s) } else { dst = strconv.AppendInt(dst, int64(%s), 10) }", ind, e.Ref, val, val)
	}
}

func emitMapKeyJSON(x *w, ind string, e Elem, val string) {
	switch e.Kind {
	case "string":
		x.p("%sdst = jsonString(dst, %s)", ind, val)
	case "int64", "uint64", "int32", "uint32":
		// JSON object keys are always strings.
		x.p("%sdst = append(dst, '\"'); dst = strconv.AppendInt(dst, int64(%s), 10); dst = append(dst, '\"')", ind, val)
	case "bool":
		x.p("%sdst = append(dst, '\"'); dst = strconv.AppendBool(dst, %s); dst = append(dst, '\"')", ind, val)
	default:
		x.p("%sdst = jsonString(dst, %s)", ind, val)
	}
}

func emitMsgJSON(x *w, ind string, e Elem, val string) {
	if e.External {
		x.p("%sdst = extJSON_%s(dst, %s)", ind, mangle(e.Ref), val)
		return
	}
	x.p("%sif %s != nil {", ind, val)
	x.p("%s\tdst, err = %s.AppendJSON(dst)", ind, val)
	x.p("%s\tif err != nil { return dst, err }", ind)
	x.p("%s} else { dst = append(dst, \"null\"...) }", ind)
}
