// embedpb is a stringer-style code generator that emits a reflection-free,
// TinyGo-safe variant of protoc-gen-go output. It reads the checked-in *.pb.go
// of a package (no protoc needed), and writes a build-tagged companion file
// whose messages carry hand-written protoiface.Methods so that
// google.golang.org/protobuf takes its fast path and never builds the
// reflection coder (the TinyGo wall).
//
// Usage:
//
//	//go:generate go run github.com/soypat/embedpb/cmd/embedpb -tag tinygo .
//	embedpb -tag js -out foo.embedded.go ./path/to/pkg
//
// Flags:
//
//	-tag     build constraint for the generated file (required). The stock
//	         .pb.go files are stamped with the negation unless -stamp-stock=false.
//	-out     output file (default <pkgdir>/embedpb_generated.go)
//	-type    comma-separated message names to include (default: all)
//	-stamp-stock  also stamp //go:build !<tag> onto stock .pb.go (default true)
//	-runtime import github.com/soypat/embedpb/pbruntime instead of emitting the
//	         per-package support code
package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/soypat/embedpb/genpb"
)

func main() {
	tag := flag.String("tag", "", "build constraint for the generated file (required)")
	out := flag.String("out", "", "output file (default <pkgdir>/embedpb_generated.go)")
	only := flag.String("type", "", "comma-separated message names to include (default all)")
	stamp := flag.Bool("stamp-stock", true, "stamp //go:build !<tag> onto stock .pb.go")
	runtime := flag.Bool("runtime", false, "import github.com/soypat/embedpb/pbruntime instead of emitting support code")
	flag.Parse()

	if *tag == "" {
		fmt.Fprintln(os.Stderr, "embedpb: -tag is required")
		os.Exit(2)
	}
	dir := "."
	if flag.NArg() > 0 {
		dir = flag.Arg(0)
	}
	var types []string
	if *only != "" {
		for n := range strings.SplitSeq(*only, ",") {
			types = append(types, strings.TrimSpace(n))
		}
	}

	if err := genpb.Generate(genpb.Options{
		Dir:        dir,
		Tag:        *tag,
		Out:        *out,
		Only:       types,
		StampStock: *stamp,
		Runtime:    *runtime,
	}); err != nil {
		fmt.Fprintln(os.Stderr, "embedpb:", err)
		os.Exit(1)
	}
}
