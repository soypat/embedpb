# embedpb

Reflection-free, TinyGo-safe protobuf codegen from existing `.pb.go` files.

`google.golang.org/protobuf` builds its wire coder by reflecting over the
generated structs, which needs `(reflect.Type).MethodByName` — unimplemented in
TinyGo's `reflect`, so any marshal panics:

```
panic: unimplemented: (reflect.Type).MethodByName()
```

embedpb reads a package's checked-in `*.pb.go` (no protoc, no `.proto`) and
writes a build-tagged companion file: replacement types carrying hand-written
`protoiface.Methods`, so the protobuf runtime takes its fast path and never
builds the reflection coder. Messages also get `AppendJSON`/`MarshalJSON`,
matching `protojson` with `UseProtoNames` + `EmitUnpopulated`, so the JSON path
sheds reflection too.

Handled: all scalar kinds, enums, nested and repeated messages, packed and
unpacked repeated, proto3 `optional`, oneofs, maps, unknown-field preservation,
and the `timestamppb`/`durationpb` well-known types.

## Usage

```sh
go run github.com/soypat/embedpb/cmd/embedpb -tag tinygo ./path/to/proto
```

```go
//go:generate go run github.com/soypat/embedpb/cmd/embedpb -tag tinygo .
```

| flag | default | meaning |
|---|---|---|
| `-tag` | required | build constraint on the generated file, e.g. `tinygo` |
| `-out` | `<pkgdir>/embedpb_generated.go` | output file |
| `-type` | all | comma-separated message names to include |
| `-stamp-stock` | `true` | also stamp `//go:build !<tag>` onto the stock `.pb.go` |
| `-runtime` | `false` | import [`pbruntime`](./pbruntime) instead of emitting support code |

As a library, for callers driving generation from their own tooling:

```go
import "github.com/soypat/embedpb/genpb"

err := genpb.Generate(genpb.Options{Dir: "./proto", Tag: "tinygo", StampStock: true})
```

`Options.Tag` may be empty in library use, which emits no build constraint —
useful when the generated package stands on its own.

## The deployment pattern

The two files must not collide in one build, so they are tagged apart:

- `embedpb_generated.go` gets `//go:build tinygo` and defines the messages.
- The stock message-bearing `.pb.go` gets `//go:build !tinygo` (that is the
  stamping; it is idempotent, so re-running is safe).
- `*_grpc.pb.go` is left untouched — service stubs define no messages and belong
  to **both** builds.

Everything under the tag is inert for normal builds, so a native `go build ./...`
is unaffected.

Either commit the generated file, or gitignore it and regenerate before the
tagged build (~1.5s per package) with `-stamp-stock=false`, since the stamps are
already committed. One caveat when retrofitting a real tree: any code path that
calls `proto.Message.String()` reaches `prototext`, which builds the reflection
coder — the same wall, from the error path.

## `-runtime`

Without it, each generated file carries its own copy of the shim
(`protoiface.Methods`, the `protoreflect.Message` adapter, JSON helpers) and the
package depends on nothing but `google.golang.org/protobuf`. With it, that code
comes from `github.com/soypat/embedpb/pbruntime` instead: ~150 lines smaller per
package, at the cost of a module dependency in whatever compiles the output.
Wire bytes and JSON are identical either way — `internal/fixture/embedded` and
`internal/fixture/embeddedrt` are the same fixture generated both ways, pinned
to the same golden bytes.

## Testing

`go test ./...` regenerates the fixtures from `internal/fixture/stock` and fails
if the checked-in output drifted; the fixture packages then round-trip against
the stock protobuf runtime in both directions and compare JSON against
`protojson`. Refresh the fixtures after an emitter change with:

```sh
go test ./genpb -run TestFixturesUpToDate -update
```

A worked end-to-end deployment — netbird's browser client under TinyGo, gRPC
over a wasm loopback, byte-identical JSON on both legs — lives in
[netbird-ssh-tests/netbird-mock](https://github.com/soypat/netbird-ssh-tests).
