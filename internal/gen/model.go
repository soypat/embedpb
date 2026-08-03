package gen

// Intermediate model of a proto package, derived from the stock *.pb.go via
// go/types. Everything the emitter needs comes from here.

// Elem classifies the leaf Go/proto type of a field (or map key/value).
type Elem struct {
	// Kind is one of: bool int32 int64 uint32 uint64 float32 float64
	// string bytes enum message
	Kind string
	// Ref is the referenced Go type name for enum/message kinds. For local
	// types it is the bare name; for external (cross-package) messages it is
	// the package-qualified name as written in source (e.g. timestamppb.Timestamp).
	Ref string
	// External is true for cross-package messages (well-known types). Their
	// codec is emitted as standalone ext* helper funcs operating on public
	// fields, since the stock type carries no reflection-free methods.
	External bool
}

// Card is a field's cardinality.
type Card int

const (
	CardSingle   Card = iota // proto3 implicit-presence scalar/message
	CardOptional             // proto3 optional (pointer scalar / nilable bytes)
	CardRepeated
	CardMap
)

type Field struct {
	GoName    string // struct field name
	ProtoName string // proto field name (protojson key with UseProtoNames)
	Num       int
	Wire      string // varint zigzag32 zigzag64 fixed32 fixed64 bytes
	Card      Card
	Elem      Elem // element type (single/optional/repeated element)

	// Map only.
	KeyElem Elem
	KeyWire string
	ValElem Elem
	ValWire string

	// Oneof only: set when this field is a oneof group (Elem/Wire unused).
	Oneof *Oneof
}

// Oneof is a real (interface-backed) oneof group.
type Oneof struct {
	GoName  string // the message struct field name (e.g. "Choice")
	IfaceGo string // interface type name (e.g. "isShapes_Choice")
	Cases   []OneofCase
}

type OneofCase struct {
	WrapperGo string // wrapper struct type name (e.g. "Shapes_ChoiceMsg")
	FieldGo   string // inner field name (e.g. "ChoiceMsg")
	ProtoName string // proto field name (protojson key)
	Num       int
	Wire      string
	Elem      Elem
}

type Message struct {
	GoName string
	Fields []Field // in field-number order (oneofs appear at their group)
}

type EnumVal struct {
	GoName    string
	ProtoName string // proto value name (for protojson), e.g. "OFFER"
	Num       int32
}

type Enum struct {
	GoName string
	Vals   []EnumVal
}

// ExtMsg is a cross-package (well-known) message whose reflection-free codec is
// emitted as standalone helper funcs operating on the stock type's public fields.
type ExtMsg struct {
	Ref    string // package-qualified Go type name, e.g. "timestamppb.Timestamp"
	Mangle string // ext-helper name suffix, e.g. "timestamppb_Timestamp"
	Fields []Field
}

type Import struct {
	Path string
	Name string
}

type Package struct {
	Name     string
	Enums    []Enum
	Messages []Message
	// Wrapper types to emit (oneof case structs + marker methods), collected
	// so the emitter can render them once.
	Oneofs    []Oneof
	Externals []ExtMsg
	Imports   []Import // extra imports for external message types
}
