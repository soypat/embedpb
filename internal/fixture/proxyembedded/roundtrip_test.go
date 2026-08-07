package proxypb

import (
	"encoding/json"
	"reflect"
	"testing"
	"time"

	stock "github.com/soypat/embedpb/internal/fixture/proxystock"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func sampleMapping() *ProxyMapping {
	return &ProxyMapping{
		Id:        "mapping-1",
		AccountId: "account-1",
		Domain:    "app.example.test",
		Path: []*PathMapping{{
			Path:   "/",
			Target: "http://upstream:8080",
			Options: &PathTargetOptions{
				SkipTlsVerify:      true,
				RequestTimeout:     durationpb.New(5 * time.Second),
				PathRewrite:        PathRewriteMode_PATH_REWRITE_PRESERVE,
				CustomHeaders:      map[string]string{"x-proxy": "netbird", "x-env": "test"},
				SessionIdleTimeout: durationpb.New(30 * time.Second),
			},
		}},
		Auth: &Authentication{
			SessionKey:           "session-key",
			MaxSessionAgeSeconds: 3600,
			Password:             true,
			Pin:                  true,
			HeaderAuths: []*HeaderAuth{{
				Header:      "Authorization",
				HashedValue: "argon2id",
			}},
		},
		PassHostHeader: true,
		ListenPort:     8443,
	}
}

func stockMapping() *stock.ProxyMapping {
	return &stock.ProxyMapping{
		Id:        "mapping-1",
		AccountId: "account-1",
		Domain:    "app.example.test",
		Path: []*stock.PathMapping{{
			Path:   "/",
			Target: "http://upstream:8080",
			Options: &stock.PathTargetOptions{
				SkipTlsVerify:      true,
				RequestTimeout:     durationpb.New(5 * time.Second),
				PathRewrite:        stock.PathRewriteMode_PATH_REWRITE_PRESERVE,
				CustomHeaders:      map[string]string{"x-proxy": "netbird", "x-env": "test"},
				SessionIdleTimeout: durationpb.New(30 * time.Second),
			},
		}},
		Auth: &stock.Authentication{
			SessionKey:           "session-key",
			MaxSessionAgeSeconds: 3600,
			Password:             true,
			Pin:                  true,
			HeaderAuths: []*stock.HeaderAuth{{
				Header:      "Authorization",
				HashedValue: "argon2id",
			}},
		},
		PassHostHeader: true,
		ListenPort:     8443,
	}
}

func TestProxyMappingInterop(t *testing.T) {
	ours, err := proto.Marshal(sampleMapping())
	if err != nil {
		t.Fatal(err)
	}
	var st stock.ProxyMapping
	if err := proto.Unmarshal(ours, &st); err != nil {
		t.Fatalf("stock unmarshal: %v", err)
	}
	if st.GetPath()[0].GetOptions().GetCustomHeaders()["x-proxy"] != "netbird" ||
		st.GetAuth().GetHeaderAuths()[0].GetHeader() != "Authorization" ||
		st.GetListenPort() != 8443 {
		t.Fatalf("stock decoded mapping wrong: %+v", &st)
	}

	stockBytes, err := proto.Marshal(stockMapping())
	if err != nil {
		t.Fatal(err)
	}
	var mine ProxyMapping
	if err := proto.Unmarshal(stockBytes, &mine); err != nil {
		t.Fatalf("our unmarshal: %v", err)
	}
	if mine.GetPath()[0].GetOptions().GetCustomHeaders()["x-env"] != "test" ||
		mine.GetAuth().GetHeaderAuths()[0].GetHashedValue() != "argon2id" ||
		mine.GetListenPort() != 8443 {
		t.Fatalf("we decoded stock mapping wrong: %+v", &mine)
	}
}

func TestProxyJSONMatchesProtojson(t *testing.T) {
	ours, err := sampleMapping().AppendJSON(nil)
	if err != nil {
		t.Fatal(err)
	}
	theirs, err := protojson.MarshalOptions{
		EmitUnpopulated: true,
		UseProtoNames:   true,
		AllowPartial:    true,
	}.Marshal(stockMapping())
	if err != nil {
		t.Fatal(err)
	}
	var gotV, wantV any
	if err := json.Unmarshal(ours, &gotV); err != nil {
		t.Fatalf("our JSON is invalid: %v\n%s", err, ours)
	}
	if err := json.Unmarshal(theirs, &wantV); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(gotV, wantV) {
		t.Fatalf("JSON mismatch\n got  %s\n want %s", ours, theirs)
	}
}

func TestProxyOneofsAndOptionalMessage(t *testing.T) {
	status := &SendStatusUpdateRequest{
		ServiceId:       "svc-1",
		AccountId:       "account-1",
		Status:          ProxyStatus_PROXY_STATUS_ERROR,
		ErrorMessage:    proto.String("certificate failed"),
		InboundListener: &ProxyInboundListener{TunnelIp: "100.64.0.10", HttpsPort: 443, HttpPort: 80},
	}
	b, err := proto.Marshal(status)
	if err != nil {
		t.Fatal(err)
	}
	var decoded SendStatusUpdateRequest
	if err := proto.Unmarshal(b, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.GetErrorMessage() != "certificate failed" || decoded.GetInboundListener().GetHttpsPort() != 443 {
		t.Fatalf("decoded optional fields wrong: %+v", &decoded)
	}

	for _, msg := range []*AuthenticateRequest{
		{Request: &AuthenticateRequest_Password{Password: &PasswordRequest{Password: "secret"}}},
		{Request: &AuthenticateRequest_Pin{Pin: &PinRequest{Pin: "123456"}}},
		{Request: &AuthenticateRequest_HeaderAuth{HeaderAuth: &HeaderAuthRequest{HeaderName: "Authorization", HeaderValue: "Bearer token"}}},
	} {
		b, err := proto.Marshal(msg)
		if err != nil {
			t.Fatal(err)
		}
		var out AuthenticateRequest
		if err := proto.Unmarshal(b, &out); err != nil {
			t.Fatal(err)
		}
		if out.GetRequest() == nil {
			t.Fatalf("decoded nil oneof for %#v", msg)
		}
	}
}

func TestProxyTimestamp(t *testing.T) {
	log := &AccessLog{
		Timestamp:  timestamppb.New(time.Unix(1800000000, 123000000).UTC()),
		LogId:      "log-1",
		ServiceId:  "svc-1",
		DurationMs: 42,
		Metadata:   map[string]string{"scenario": "fixture"},
	}
	b, err := proto.Marshal(log)
	if err != nil {
		t.Fatal(err)
	}
	var out AccessLog
	if err := proto.Unmarshal(b, &out); err != nil {
		t.Fatal(err)
	}
	if out.GetTimestamp().GetNanos() != 123000000 || out.GetMetadata()["scenario"] != "fixture" {
		t.Fatalf("decoded access log wrong: %+v", &out)
	}
}
