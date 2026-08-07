package proxypb

import (
	"testing"
	"time"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/durationpb"
)

func TestRuntimeProxyMapping(t *testing.T) {
	msg := &ProxyMapping{
		Id:        "mapping-rt",
		AccountId: "account-rt",
		Domain:    "runtime.example.test",
		Path: []*PathMapping{{
			Path:   "/api",
			Target: "http://runtime:8080",
			Options: &PathTargetOptions{
				RequestTimeout: durationpb.New(2 * time.Second),
				CustomHeaders:  map[string]string{"x-runtime": "true"},
			},
		}},
	}
	b, err := proto.Marshal(msg)
	if err != nil {
		t.Fatal(err)
	}
	var out ProxyMapping
	if err := proto.Unmarshal(b, &out); err != nil {
		t.Fatal(err)
	}
	if out.GetPath()[0].GetOptions().GetCustomHeaders()["x-runtime"] != "true" {
		t.Fatalf("decoded runtime mapping wrong: %+v", &out)
	}
}
