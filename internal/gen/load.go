package gen

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"reflect"
	"sort"
	"strconv"
	"strings"

	"golang.org/x/tools/go/packages"
)

// load parses the stock *.pb.go of dir via go/types and builds a Package model.
func load(dir string, only map[string]bool) (*Package, error) {
	cfg := &packages.Config{
		Mode: packages.NeedName | packages.NeedTypes | packages.NeedTypesInfo |
			packages.NeedSyntax | packages.NeedImports | packages.NeedDeps,
		Dir: dir,
	}
	pkgs, err := packages.Load(cfg, ".")
	if err != nil {
		return nil, err
	}
	if len(pkgs) != 1 {
		return nil, fmt.Errorf("expected 1 package in %s, got %d", dir, len(pkgs))
	}
	p := pkgs[0]
	if len(p.Errors) > 0 {
		return nil, fmt.Errorf("load errors: %v", p.Errors)
	}
	scope := p.Types.Scope()

	out := &Package{Name: p.Name}

	// Pass 1: index enums (named integer types) and their values, and collect
	// oneof marker interfaces + message struct types.
	enumByName := map[string]*Enum{}
	var enumOrder []string
	ifaceTypes := map[string]*types.Named{} // "isFoo_Bar" -> named iface
	msgNamed := map[string]*types.Named{}   // message struct types

	for _, name := range scope.Names() {
		obj := scope.Lookup(name)
		tn, ok := obj.(*types.TypeName)
		if !ok {
			continue
		}
		named, ok := tn.Type().(*types.Named)
		if !ok {
			continue
		}
		switch u := named.Underlying().(type) {
		case *types.Basic:
			if u.Info()&types.IsInteger != 0 {
				enumByName[name] = &Enum{GoName: name}
				enumOrder = append(enumOrder, name)
			}
		case *types.Interface:
			if isOneofIface(name) {
				ifaceTypes[name] = named
			}
		case *types.Struct:
			if hasMethod(named, "ProtoReflect") {
				msgNamed[name] = named
			}
		}
	}

	// proto value names from the generated `<Enum>_name` maps (for protojson).
	protoNames := enumProtoNames(p)

	// enum values from package-scope constants.
	for _, name := range scope.Names() {
		c, ok := scope.Lookup(name).(*types.Const)
		if !ok {
			continue
		}
		named, ok := c.Type().(*types.Named)
		if !ok {
			continue
		}
		enumName := named.Obj().Name()
		e, ok := enumByName[enumName]
		if !ok {
			continue
		}
		iv, ok := constInt32(c)
		if !ok {
			continue
		}
		pn := name
		if m, ok := protoNames[enumName]; ok {
			if s, ok := m[iv]; ok {
				pn = s
			}
		}
		e.Vals = append(e.Vals, EnumVal{GoName: name, ProtoName: pn, Num: iv})
	}
	for _, n := range enumOrder {
		e := enumByName[n]
		if len(e.Vals) == 0 {
			continue
		}
		sort.Slice(e.Vals, func(i, j int) bool { return e.Vals[i].Num < e.Vals[j].Num })
		out.Enums = append(out.Enums, *e)
	}

	ext := map[string]*types.Named{} // external message Ref -> named

	// oneof case wrappers: for each marker iface, find implementing structs.
	oneofCases := map[string][]OneofCase{} // iface name -> cases
	for ifaceName, ifaceNamed := range ifaceTypes {
		ifaceT := ifaceNamed.Underlying().(*types.Interface)
		var cases []OneofCase
		for wName, wNamed := range allStructs(scope) {
			if !types.Implements(types.NewPointer(wNamed), ifaceT) {
				continue
			}
			st := wNamed.Underlying().(*types.Struct)
			// wrapper has exactly one protobuf-tagged field.
			for i := 0; i < st.NumFields(); i++ {
				tag, ok := parseProtoTag(reflect.StructTag(st.Tag(i)).Get("protobuf"))
				if !ok {
					continue
				}
				f := st.Field(i)
				elem, _, err := classifyElem(f.Type(), tag, msgNamed, enumByName, ext)
				if err != nil {
					return nil, fmt.Errorf("oneof %s case %s: %w", ifaceName, wName, err)
				}
				cases = append(cases, OneofCase{
					WrapperGo: wName,
					FieldGo:   f.Name(),
					ProtoName: tag.Name,
					Num:       tag.Num,
					Wire:      tag.Wire,
					Elem:      elem,
				})
			}
		}
		sort.Slice(cases, func(i, j int) bool { return cases[i].Num < cases[j].Num })
		oneofCases[ifaceName] = cases
	}

	// Pass 2: messages.
	var msgNames []string
	for n := range msgNamed {
		if only == nil || only[n] {
			msgNames = append(msgNames, n)
		}
	}
	sort.Strings(msgNames)

	seenOneof := map[string]bool{}
	for _, name := range msgNames {
		named := msgNamed[name]
		st := named.Underlying().(*types.Struct)
		msg := Message{GoName: name}
		for i := 0; i < st.NumFields(); i++ {
			f := st.Field(i)
			rawtag := reflect.StructTag(st.Tag(i))
			if oneofName := rawtag.Get("protobuf_oneof"); oneofName != "" {
				// real oneof group: field type is a marker interface.
				ifaceNamed, ok := f.Type().(*types.Named)
				if !ok {
					return nil, fmt.Errorf("%s.%s: oneof field not a named iface", name, f.Name())
				}
				ifn := ifaceNamed.Obj().Name()
				cases := oneofCases[ifn]
				on := Oneof{GoName: f.Name(), IfaceGo: ifn, Cases: cases}
				minNum := 1 << 30
				for _, c := range cases {
					if c.Num < minNum {
						minNum = c.Num
					}
				}
				msg.Fields = append(msg.Fields, Field{Num: minNum, Oneof: &on})
				if !seenOneof[ifn] {
					out.Oneofs = append(out.Oneofs, on)
					seenOneof[ifn] = true
				}
				continue
			}
			tag, ok := parseProtoTag(rawtag.Get("protobuf"))
			if !ok {
				continue // internal state field
			}
			fld, err := classifyField(f, tag, msgNamed, enumByName, ext)
			if err != nil {
				return nil, fmt.Errorf("%s.%s: %w", name, f.Name(), err)
			}
			msg.Fields = append(msg.Fields, fld)
		}
		sort.SliceStable(msg.Fields, func(i, j int) bool { return msg.Fields[i].Num < msg.Fields[j].Num })
		out.Messages = append(out.Messages, msg)
	}

	// Build external (well-known) message models + imports. New externals may
	// surface while classifying an external's own fields, so use a worklist.
	imports := map[string]string{} // path -> name
	built := map[string]bool{}
	for {
		var ref string
		var named *types.Named
		for r, n := range ext {
			if !built[r] {
				ref, named = r, n
				break
			}
		}
		if named == nil {
			break
		}
		built[ref] = true
		if pkg := named.Obj().Pkg(); pkg != nil {
			imports[pkg.Path()] = pkg.Name()
		}
		st, ok := named.Underlying().(*types.Struct)
		if !ok {
			return nil, fmt.Errorf("external %s not a struct", ref)
		}
		em := ExtMsg{Ref: ref, Mangle: mangle(ref)}
		for i := 0; i < st.NumFields(); i++ {
			f := st.Field(i)
			tag, ok := parseProtoTag(reflect.StructTag(st.Tag(i)).Get("protobuf"))
			if !ok {
				continue
			}
			fld, err := classifyField(f, tag, msgNamed, enumByName, ext)
			if err != nil {
				return nil, fmt.Errorf("external %s.%s: %w", ref, f.Name(), err)
			}
			em.Fields = append(em.Fields, fld)
		}
		sort.SliceStable(em.Fields, func(i, j int) bool { return em.Fields[i].Num < em.Fields[j].Num })
		out.Externals = append(out.Externals, em)
	}
	sort.Slice(out.Externals, func(i, j int) bool { return out.Externals[i].Ref < out.Externals[j].Ref })
	for p, n := range imports {
		out.Imports = append(out.Imports, Import{Path: p, Name: n})
	}
	sort.Slice(out.Imports, func(i, j int) bool { return out.Imports[i].Path < out.Imports[j].Path })
	return out, nil
}

// enumProtoNames reads the generated `<Enum>_name = map[int32]string{...}` vars
// to recover proto enum value names (protoc-gen-go's const names are not the
// proto names, e.g. Body_OFFER vs "OFFER").
func enumProtoNames(p *packages.Package) map[string]map[int32]string {
	out := map[string]map[int32]string{}
	for _, f := range p.Syntax {
		for _, d := range f.Decls {
			gd, ok := d.(*ast.GenDecl)
			if !ok || gd.Tok != token.VAR {
				continue
			}
			for _, s := range gd.Specs {
				vs, ok := s.(*ast.ValueSpec)
				if !ok || len(vs.Names) != 1 || len(vs.Values) != 1 {
					continue
				}
				name := vs.Names[0].Name
				if !strings.HasSuffix(name, "_name") {
					continue
				}
				cl, ok := vs.Values[0].(*ast.CompositeLit)
				if !ok {
					continue
				}
				m := map[int32]string{}
				for _, el := range cl.Elts {
					kv, ok := el.(*ast.KeyValueExpr)
					if !ok {
						continue
					}
					klit, ok := kv.Key.(*ast.BasicLit)
					if !ok || klit.Kind != token.INT {
						continue
					}
					vlit, ok := kv.Value.(*ast.BasicLit)
					if !ok || vlit.Kind != token.STRING {
						continue
					}
					n, err := strconv.Atoi(klit.Value)
					if err != nil {
						continue
					}
					sv, err := strconv.Unquote(vlit.Value)
					if err != nil {
						continue
					}
					m[int32(n)] = sv
				}
				if len(m) > 0 {
					out[strings.TrimSuffix(name, "_name")] = m
				}
			}
		}
	}
	return out
}

func mangle(ref string) string {
	out := make([]byte, len(ref))
	for i := 0; i < len(ref); i++ {
		if ref[i] == '.' {
			out[i] = '_'
		} else {
			out[i] = ref[i]
		}
	}
	return string(out)
}

func classifyField(f *types.Var, tag protoTag, msgNamed map[string]*types.Named, enums map[string]*Enum, ext map[string]*types.Named) (Field, error) {
	fld := Field{GoName: f.Name(), ProtoName: tag.Name, Num: tag.Num, Wire: tag.Wire}
	t := f.Type()

	if mt, ok := t.(*types.Map); ok {
		fld.Card = CardMap
		kElem, kWire, err := elemFromType(mt.Key(), msgNamed, enums, ext)
		if err != nil {
			return fld, fmt.Errorf("map key: %w", err)
		}
		vElem, vWire, err := elemFromType(mt.Elem(), msgNamed, enums, ext)
		if err != nil {
			return fld, fmt.Errorf("map val: %w", err)
		}
		fld.KeyElem, fld.KeyWire = kElem, kWire
		fld.ValElem, fld.ValWire = vElem, vWire
		return fld, nil
	}

	if sl, ok := t.(*types.Slice); ok {
		if isByte(sl.Elem()) { // []byte scalar/bytes
			fld.Elem = Elem{Kind: "bytes"}
			if tag.Oneof {
				fld.Card = CardOptional
			} else {
				fld.Card = CardSingle
			}
			return fld, nil
		}
		// repeated element
		elem, _, err := elemFromType(sl.Elem(), msgNamed, enums, ext)
		if err != nil {
			return fld, err
		}
		fld.Elem = elem
		fld.Card = CardRepeated
		return fld, nil
	}

	if pt, ok := t.(*types.Pointer); ok {
		if named, ok := pt.Elem().(*types.Named); ok {
			if _, isMsg := msgNamed[named.Obj().Name()]; isMsg {
				fld.Elem = Elem{Kind: "message", Ref: named.Obj().Name()}
				fld.Card = CardSingle
				return fld, nil
			}
			// pointer to external message (well-known type) => single message.
			if isExternalMessage(named) {
				fld.Elem = registerExternal(named, ext)
				fld.Card = CardSingle
				return fld, nil
			}
		}
		// pointer to scalar/enum => proto3 optional
		elem, _, err := elemFromType(pt.Elem(), msgNamed, enums, ext)
		if err != nil {
			return fld, err
		}
		fld.Elem = elem
		fld.Card = CardOptional
		return fld, nil
	}

	elem, _, err := elemFromType(t, msgNamed, enums, ext)
	if err != nil {
		return fld, err
	}
	fld.Elem = elem
	fld.Card = CardSingle
	return fld, nil
}

// classifyElem is like classifyField but only returns the Elem (for oneof cases).
func classifyElem(t types.Type, tag protoTag, msgNamed map[string]*types.Named, enums map[string]*Enum, ext map[string]*types.Named) (Elem, string, error) {
	if pt, ok := t.(*types.Pointer); ok {
		if named, ok := pt.Elem().(*types.Named); ok {
			if _, isMsg := msgNamed[named.Obj().Name()]; isMsg {
				return Elem{Kind: "message", Ref: named.Obj().Name()}, "bytes", nil
			}
			if isExternalMessage(named) {
				return registerExternal(named, ext), "bytes", nil
			}
		}
		return elemFromType(pt.Elem(), msgNamed, enums, ext)
	}
	if sl, ok := t.(*types.Slice); ok && isByte(sl.Elem()) {
		return Elem{Kind: "bytes"}, "bytes", nil
	}
	return elemFromType(t, msgNamed, enums, ext)
}

// elemFromType maps a leaf Go type to an Elem plus its natural wire kind.
func elemFromType(t types.Type, msgNamed map[string]*types.Named, enums map[string]*Enum, ext map[string]*types.Named) (Elem, string, error) {
	switch tt := t.(type) {
	case *types.Pointer:
		return elemFromType(tt.Elem(), msgNamed, enums, ext)
	case *types.Named:
		nm := tt.Obj().Name()
		if _, ok := msgNamed[nm]; ok {
			return Elem{Kind: "message", Ref: nm}, "bytes", nil
		}
		if _, ok := enums[nm]; ok {
			return Elem{Kind: "enum", Ref: nm}, "varint", nil
		}
		if isExternalMessage(tt) {
			return registerExternal(tt, ext), "bytes", nil
		}
		return Elem{}, "", fmt.Errorf("unsupported named type %q (cross-package?)", nm)
	case *types.Slice:
		if isByte(tt.Elem()) {
			return Elem{Kind: "bytes"}, "bytes", nil
		}
		return Elem{}, "", fmt.Errorf("nested slice unsupported")
	case *types.Basic:
		switch tt.Kind() {
		case types.Bool:
			return Elem{Kind: "bool"}, "varint", nil
		case types.Int32:
			return Elem{Kind: "int32"}, "varint", nil
		case types.Int64:
			return Elem{Kind: "int64"}, "varint", nil
		case types.Uint32:
			return Elem{Kind: "uint32"}, "varint", nil
		case types.Uint64:
			return Elem{Kind: "uint64"}, "varint", nil
		case types.Float32:
			return Elem{Kind: "float32"}, "fixed32", nil
		case types.Float64:
			return Elem{Kind: "float64"}, "fixed64", nil
		case types.String:
			return Elem{Kind: "string"}, "bytes", nil
		}
	}
	return Elem{}, "", fmt.Errorf("unsupported type %s", t)
}

// --- go/types helpers ---

// isExternalMessage reports whether named is a proto message from another
// package (a well-known type). Detected by the ProtoReflect method.
func isExternalMessage(named *types.Named) bool {
	if _, ok := named.Underlying().(*types.Struct); !ok {
		return false
	}
	return hasMethod(named, "ProtoReflect")
}

// registerExternal records an external message and returns its Elem (with the
// package-qualified Ref used verbatim in generated source).
func registerExternal(named *types.Named, ext map[string]*types.Named) Elem {
	pkg := named.Obj().Pkg()
	ref := named.Obj().Name()
	if pkg != nil {
		ref = pkg.Name() + "." + ref
	}
	ext[ref] = named
	return Elem{Kind: "message", Ref: ref, External: true}
}

func isOneofIface(name string) bool { return len(name) > 2 && name[:2] == "is" }

func isByte(t types.Type) bool {
	b, ok := t.Underlying().(*types.Basic)
	return ok && b.Kind() == types.Uint8
}

func hasMethod(named *types.Named, name string) bool {
	pt := types.NewPointer(named)
	ms := types.NewMethodSet(pt)
	for method := range ms.Methods() {
		if method.Obj().Name() == name {
			return true
		}
	}
	return false
}

func allStructs(scope *types.Scope) map[string]*types.Named {
	out := map[string]*types.Named{}
	for _, n := range scope.Names() {
		tn, ok := scope.Lookup(n).(*types.TypeName)
		if !ok {
			continue
		}
		named, ok := tn.Type().(*types.Named)
		if !ok {
			continue
		}
		if _, ok := named.Underlying().(*types.Struct); ok {
			out[n] = named
		}
	}
	return out
}

func constInt32(c *types.Const) (int32, bool) {
	// c.Val() is a constant.Value of kind Int.
	i, ok := constToInt64(c)
	if !ok {
		return 0, false
	}
	return int32(i), true
}
