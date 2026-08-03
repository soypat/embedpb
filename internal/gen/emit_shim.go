package gen

// renderShim emits the reflection-free protoreflect.Message shim, the
// hand-written protoiface.Methods that keep proto.Marshal on its fast path, and
// the JSON helpers. Emitted once per generated file.
//
// With runtime set the same names become aliases onto pbruntime instead, so the
// message bodies — which only ever refer to msgReflect, embedMethods, embedBool
// and the json* helpers — are emitted identically in both modes.
func renderShim(x *w, runtime bool) {
	if runtime {
		x.p("%s", shimAliases)
		return
	}
	x.p("%s", shimInline)
}

const shimAliases = `
type msgReflect = pbruntime.Message

var embedMethods = pbruntime.Methods

func embedBool(v bool) uint64                { return pbruntime.Bool(v) }
func jsonEndObj(dst []byte) []byte           { return pbruntime.JSONEndObj(dst) }
func jsonString(dst []byte, s string) []byte { return pbruntime.JSONString(dst, s) }
func jsonBytes(dst []byte, b []byte) []byte  { return pbruntime.JSONBytes(dst, b) }
func jsonNanoFrac(nanos int32) string        { return pbruntime.JSONNanoFrac(nanos) }
`

const shimInline = `
func embedBool(v bool) uint64 {
	if v {
		return 1
	}
	return 0
}

type msgReflect struct {
	MarshalFn    func([]byte) ([]byte, error)
	SizeFn       func() int
	UnmarshalFn  func([]byte) error
	Iface        protoreflect.ProtoMessage
	Valid        bool
	NewFn        func() protoreflect.Message
	GetUnknownFn func() protoreflect.RawFields
	SetUnknownFn func(protoreflect.RawFields)
}

var _ protoreflect.Message = msgReflect{}

func (r msgReflect) ProtoMethods() *protoiface.Methods    { return &embedMethods }
func (r msgReflect) Interface() protoreflect.ProtoMessage { return r.Iface }
func (r msgReflect) IsValid() bool                        { return r.Valid }
func (r msgReflect) New() protoreflect.Message            { return r.NewFn() }
func (r msgReflect) GetUnknown() protoreflect.RawFields   { return r.GetUnknownFn() }
func (r msgReflect) SetUnknown(f protoreflect.RawFields)  { r.SetUnknownFn(f) }

func embedCanary(m string) string {
	return "CANARY: protoreflect.Message." + m + " called on reflection-free message"
}

func (r msgReflect) Descriptor() protoreflect.MessageDescriptor { panic(embedCanary("Descriptor")) }
func (r msgReflect) Type() protoreflect.MessageType             { panic(embedCanary("Type")) }
func (r msgReflect) Range(func(protoreflect.FieldDescriptor, protoreflect.Value) bool) {
	panic(embedCanary("Range"))
}
func (r msgReflect) Has(protoreflect.FieldDescriptor) bool               { panic(embedCanary("Has")) }
func (r msgReflect) Clear(protoreflect.FieldDescriptor)                  { panic(embedCanary("Clear")) }
func (r msgReflect) Get(protoreflect.FieldDescriptor) protoreflect.Value { panic(embedCanary("Get")) }
func (r msgReflect) Set(protoreflect.FieldDescriptor, protoreflect.Value) {
	panic(embedCanary("Set"))
}
func (r msgReflect) Mutable(protoreflect.FieldDescriptor) protoreflect.Value {
	panic(embedCanary("Mutable"))
}
func (r msgReflect) NewField(protoreflect.FieldDescriptor) protoreflect.Value {
	panic(embedCanary("NewField"))
}
func (r msgReflect) WhichOneof(protoreflect.OneofDescriptor) protoreflect.FieldDescriptor {
	panic(embedCanary("WhichOneof"))
}

var embedMethods = protoiface.Methods{
	Flags: protoiface.SupportMarshalDeterministic | protoiface.SupportUnmarshalDiscardUnknown,
	Size: func(in protoiface.SizeInput) protoiface.SizeOutput {
		return protoiface.SizeOutput{Size: in.Message.(msgReflect).SizeFn()}
	},
	Marshal: func(in protoiface.MarshalInput) (protoiface.MarshalOutput, error) {
		b, err := in.Message.(msgReflect).MarshalFn(in.Buf)
		return protoiface.MarshalOutput{Buf: b}, err
	},
	Unmarshal: func(in protoiface.UnmarshalInput) (protoiface.UnmarshalOutput, error) {
		err := in.Message.(msgReflect).UnmarshalFn(in.Buf)
		return protoiface.UnmarshalOutput{Flags: protoiface.UnmarshalInitialized}, err
	},
	Merge: func(in protoiface.MergeInput) protoiface.MergeOutput {
		src := in.Source.(msgReflect)
		dst := in.Destination.(msgReflect)
		sb, _ := src.MarshalFn(nil)
		if err := dst.UnmarshalFn(sb); err != nil {
			return protoiface.MergeOutput{}
		}
		return protoiface.MergeOutput{Flags: protoiface.MergeComplete}
	},
	CheckInitialized: func(in protoiface.CheckInitializedInput) (protoiface.CheckInitializedOutput, error) {
		return protoiface.CheckInitializedOutput{}, nil
	},
	Equal: func(in protoiface.EqualInput) protoiface.EqualOutput {
		a, ok1 := in.MessageA.(msgReflect)
		b, ok2 := in.MessageB.(msgReflect)
		if !ok1 || !ok2 {
			return protoiface.EqualOutput{Equal: false}
		}
		ab, _ := a.MarshalFn(nil)
		bb, _ := b.MarshalFn(nil)
		return protoiface.EqualOutput{Equal: bytes.Equal(ab, bb)}
	},
}

// --- protojson-compatible JSON helpers ---

func jsonEndObj(dst []byte) []byte {
	if n := len(dst); n > 0 && dst[n-1] == ',' {
		dst[n-1] = '}'
		return dst
	}
	return append(dst, '}')
}

func jsonHexDigit(b byte) byte {
	if b < 10 {
		return '0' + b
	}
	return 'a' + b - 10
}

func jsonString(dst []byte, s string) []byte {
	dst = append(dst, '"')
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch c {
		case '"':
			dst = append(dst, '\\', '"')
		case '\\':
			dst = append(dst, '\\', '\\')
		case '\n':
			dst = append(dst, '\\', 'n')
		case '\r':
			dst = append(dst, '\\', 'r')
		case '\t':
			dst = append(dst, '\\', 't')
		default:
			if c < 0x20 {
				dst = append(dst, '\\', 'u', '0', '0', jsonHexDigit(c>>4), jsonHexDigit(c&0xf))
			} else {
				dst = append(dst, c)
			}
		}
	}
	return append(dst, '"')
}

func jsonBytes(dst []byte, b []byte) []byte {
	dst = append(dst, '"')
	dst = base64.StdEncoding.AppendEncode(dst, b)
	return append(dst, '"')
}

// jsonNanoFrac formats sub-second nanos as protojson does: trailing zero groups
// of three trimmed, leaving 3, 6, or 9 fractional digits.
func jsonNanoFrac(nanos int32) string {
	buf := fmt.Sprintf("%09d", nanos)
	n := 9
	for n > 3 && buf[n-1] == '0' && buf[n-2] == '0' && buf[n-3] == '0' {
		n -= 3
	}
	return buf[:n]
}
`
