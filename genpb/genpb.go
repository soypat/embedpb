// Package genpb generates a reflection-free, TinyGo-safe variant of
// protoc-gen-go output.
//
// google.golang.org/protobuf builds its wire coder by reflecting over the
// generated structs, which requires (reflect.Type).MethodByName — unimplemented
// in TinyGo's reflect. genpb reads a package's checked-in *.pb.go (no protoc
// needed) and emits a companion file whose messages carry hand-written
// protoiface.Methods, so the protobuf runtime takes its fast path and never
// builds the reflection coder. It also emits a protojson-compatible AppendJSON.
//
// The command lives at cmd/embedpb; this package is the same generator as a
// library, for callers that drive generation from their own tooling.
package genpb

import "github.com/soypat/embedpb/internal/gen"

// Options configures a generation run.
type Options struct {
	// Dir is the package directory to read. Empty means the working directory.
	Dir string
	// Tag is the build constraint placed on the generated file, e.g. "tinygo"
	// emits "//go:build tinygo". Empty emits no constraint, in which case
	// StampStock must be false (there is nothing to negate).
	Tag string
	// Out is the output file. Empty means <Dir>/embedpb_generated.go.
	Out string
	// Only limits generation to the named messages. Nil means all of them.
	Only []string
	// StampStock adds "//go:build !<Tag>" to the stock message-bearing *.pb.go
	// in Dir, so they leave the build the generated file targets. Service stubs
	// (*_grpc.pb.go) are left alone: they define no messages and belong to both
	// builds. Stamping is idempotent.
	StampStock bool
	// Runtime makes the generated file import github.com/soypat/embedpb/pbruntime
	// for its support code instead of emitting a private copy of it. Smaller
	// output, at the cost of a module dependency for whoever compiles it.
	Runtime bool
}

// Generate writes the reflection-free companion file for the package in
// opts.Dir, and stamps the stock files if opts.StampStock is set.
func Generate(opts Options) error {
	return gen.Run(gen.Options{
		Dir:        opts.Dir,
		Tag:        opts.Tag,
		Out:        opts.Out,
		Only:       opts.Only,
		StampStock: opts.StampStock,
		Runtime:    opts.Runtime,
	})
}
