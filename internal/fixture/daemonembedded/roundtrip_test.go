package daemonpb

import (
	"encoding/json"
	"reflect"
	"testing"
	"time"

	stock "github.com/soypat/embedpb/internal/fixture/daemonstock"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func boolp(v bool) *bool {
	p := new(bool)
	*p = v
	return p
}

func stringp(v string) *string {
	p := new(string)
	*p = v
	return p
}

func int64p(v int64) *int64 {
	p := new(int64)
	*p = v
	return p
}

func int32p(v int32) *int32 {
	p := new(int32)
	*p = v
	return p
}

func sampleLoginRequest() *LoginRequest {
	return &LoginRequest{
		SetupKey:                      "setup-key",
		PreSharedKey:                  "legacy-key",
		ManagementUrl:                 "https://management.netbird.test",
		AdminURL:                      "https://app.netbird.test",
		NatExternalIPs:                []string{"203.0.113.10", "2001:db8::10"},
		CleanNATExternalIPs:           true,
		CustomDNSAddress:              []byte{100, 64, 0, 53},
		IsUnixDesktopClient:           true,
		Hostname:                      "peer-a",
		RosenpassEnabled:              boolp(true),
		InterfaceName:                 stringp("wg0"),
		WireguardPort:                 int64p(51820),
		OptionalPreSharedKey:          stringp("modern-key"),
		DisableAutoConnect:            boolp(true),
		ServerSSHAllowed:              boolp(true),
		RosenpassPermissive:           boolp(false),
		ExtraIFaceBlacklist:           []string{"docker0", "virbr0"},
		NetworkMonitor:                boolp(true),
		DnsRouteInterval:              durationpb.New(30 * time.Second),
		DisableClientRoutes:           boolp(false),
		DisableServerRoutes:           boolp(true),
		DisableDns:                    boolp(false),
		DisableFirewall:               boolp(true),
		BlockLanAccess:                boolp(true),
		DisableNotifications:          boolp(false),
		DnsLabels:                     []string{"prod", "ssh"},
		CleanDNSLabels:                true,
		LazyConnectionEnabled:         boolp(true),
		BlockInbound:                  boolp(true),
		ProfileName:                   stringp("work"),
		Username:                      stringp("user@example.test"),
		Mtu:                           int64p(1280),
		Hint:                          stringp("user@example.test"),
		EnableSSHRoot:                 boolp(true),
		EnableSSHSFTP:                 boolp(true),
		EnableSSHLocalPortForwarding:  boolp(true),
		EnableSSHRemotePortForwarding: boolp(false),
		DisableSSHAuth:                boolp(false),
		SshJWTCacheTTL:                int32p(600),
		DisableIpv6:                   boolp(false),
	}
}

func sampleStatusResponse() *StatusResponse {
	now := time.Unix(1800000000, 123000000).UTC()
	return &StatusResponse{
		Status:        "connected",
		DaemonVersion: "0.99.0",
		FullStatus: &FullStatus{
			ManagementState: &ManagementState{URL: "https://management.netbird.test", Connected: true},
			SignalState:     &SignalState{URL: "https://signal.netbird.test", Connected: true},
			LocalPeerState: &LocalPeerState{
				IP:                  "100.64.0.1",
				Ipv6:                "fd00::1",
				PubKey:              "self-key",
				KernelInterface:     true,
				Fqdn:                "self.netbird.test",
				RosenpassEnabled:    true,
				RosenpassPermissive: true,
				Networks:            []string{"net-a", "net-b"},
				WgPort:              51820,
			},
			Peers: []*PeerState{{
				IP:                         "100.64.0.2",
				Ipv6:                       "fd00::2",
				PubKey:                     "peer-key",
				ConnStatus:                 "connected",
				ConnStatusUpdate:           timestamppb.New(now),
				Relayed:                    true,
				LocalIceCandidateType:      "host",
				RemoteIceCandidateType:     "srflx",
				Fqdn:                       "peer.netbird.test",
				LocalIceCandidateEndpoint:  "100.64.0.1:51820",
				RemoteIceCandidateEndpoint: "198.51.100.10:51820",
				LastWireguardHandshake:     timestamppb.New(now.Add(-time.Minute)),
				BytesRx:                    1024,
				BytesTx:                    2048,
				RosenpassEnabled:           true,
				Networks:                   []string{"net-a"},
				Latency:                    durationpb.New(12 * time.Millisecond),
				RelayAddress:               "rels://relay.netbird.test",
				SshHostKey:                 []byte("ssh-ed25519 AAAAMOCK"),
			}},
			Relays: []*RelayState{{URI: "rels://relay.netbird.test", Available: true, Transport: "quic"}},
			DnsServers: []*NSGroupState{{
				Servers: []string{"100.64.0.53"},
				Domains: []string{"corp.example.test"},
				Enabled: true,
			}},
			Events: []*SystemEvent{{
				Id:          "evt-1",
				Severity:    SystemEvent_WARNING,
				Category:    SystemEvent_NETWORK,
				Message:     "route changed",
				UserMessage: "Route changed",
				Timestamp:   timestamppb.New(now),
				Metadata:    map[string]string{"route": "net-a"},
			}},
			NumberOfForwardingRules: 2,
			LazyConnectionEnabled:   true,
			SshServerState: &SSHServerState{
				Enabled: true,
				Sessions: []*SSHSessionInfo{{
					Username:      "root",
					RemoteAddress: "100.64.0.2:49200",
					Command:       "uptime",
					JwtUsername:   "user@example.test",
					PortForwards:  []string{"127.0.0.1:8080"},
				}},
			},
		},
	}
}

func sampleForwardingRules() *ForwardingRulesResponse {
	return &ForwardingRulesResponse{Rules: []*ForwardingRule{
		{
			Protocol:           "tcp",
			DestinationPort:    &PortInfo{PortSelection: &PortInfo_Port{Port: 443}},
			TranslatedAddress:  "100.64.0.10",
			TranslatedHostname: "app.netbird.test",
			TranslatedPort:     &PortInfo{PortSelection: &PortInfo_Port{Port: 8443}},
		},
		{
			Protocol:          "udp",
			DestinationPort:   &PortInfo{PortSelection: &PortInfo_Range_{Range: &PortInfo_Range{Start: 5000, End: 6000}}},
			TranslatedAddress: "100.64.0.11",
			TranslatedPort:    &PortInfo{PortSelection: &PortInfo_Range_{Range: &PortInfo_Range{Start: 15000, End: 16000}}},
		},
	}}
}

func stockLoginRequest() *stock.LoginRequest {
	var out stock.LoginRequest
	mustRoundtrip(sampleLoginRequest(), &out)
	return &out
}

func stockStatusResponse() *stock.StatusResponse {
	var out stock.StatusResponse
	mustRoundtrip(sampleStatusResponse(), &out)
	return &out
}

func mustRoundtrip(in proto.Message, out proto.Message) {
	b, err := proto.Marshal(in)
	if err != nil {
		panic(err)
	}
	if err := proto.Unmarshal(b, out); err != nil {
		panic(err)
	}
}

func TestDaemonLoginInterop(t *testing.T) {
	ours, err := proto.Marshal(sampleLoginRequest())
	if err != nil {
		t.Fatal(err)
	}
	var st stock.LoginRequest
	if err := proto.Unmarshal(ours, &st); err != nil {
		t.Fatalf("stock unmarshal: %v", err)
	}
	if !st.GetRosenpassEnabled() ||
		st.GetDnsRouteInterval().AsDuration() != 30*time.Second ||
		st.GetSshJWTCacheTTL() != 600 ||
		st.GetDnsLabels()[1] != "ssh" {
		t.Fatalf("stock decoded daemon login wrong: %+v", &st)
	}

	stockBytes, err := proto.Marshal(stockLoginRequest())
	if err != nil {
		t.Fatal(err)
	}
	var mine LoginRequest
	if err := proto.Unmarshal(stockBytes, &mine); err != nil {
		t.Fatalf("our unmarshal: %v", err)
	}
	if !mine.GetBlockInbound() ||
		mine.GetInterfaceName() != "wg0" ||
		mine.GetWireguardPort() != 51820 ||
		mine.GetExtraIFaceBlacklist()[0] != "docker0" {
		t.Fatalf("we decoded stock daemon login wrong: %+v", &mine)
	}
}

func TestDaemonForwardingInterop(t *testing.T) {
	ours, err := proto.Marshal(sampleForwardingRules())
	if err != nil {
		t.Fatal(err)
	}
	var st stock.ForwardingRulesResponse
	if err := proto.Unmarshal(ours, &st); err != nil {
		t.Fatalf("stock unmarshal: %v", err)
	}
	if st.GetRules()[0].GetDestinationPort().GetPort() != 443 ||
		st.GetRules()[1].GetDestinationPort().GetRange().GetEnd() != 6000 ||
		st.GetRules()[1].GetTranslatedPort().GetRange().GetStart() != 15000 {
		t.Fatalf("stock decoded daemon forwarding wrong: %+v", &st)
	}
}

func TestDaemonNetworkMapInterop(t *testing.T) {
	msg := &Network{
		ID:       "network-a",
		Range:    "10.10.0.0/16",
		Selected: true,
		Domains:  []string{"corp.example.test"},
		ResolvedIPs: map[string]*IPList{
			"app.corp.example.test": {Ips: []string{"100.64.0.10", "fd00::10"}},
		},
	}
	b, err := proto.Marshal(msg)
	if err != nil {
		t.Fatal(err)
	}
	var st stock.Network
	if err := proto.Unmarshal(b, &st); err != nil {
		t.Fatalf("stock unmarshal: %v", err)
	}
	if st.GetResolvedIPs()["app.corp.example.test"].GetIps()[1] != "fd00::10" {
		t.Fatalf("stock decoded daemon network wrong: %+v", &st)
	}
}

func TestDaemonStatusJSONMatchesProtojson(t *testing.T) {
	ours, err := sampleStatusResponse().AppendJSON(nil)
	if err != nil {
		t.Fatal(err)
	}
	theirs, err := protojson.MarshalOptions{
		EmitUnpopulated: true,
		UseProtoNames:   true,
		AllowPartial:    true,
	}.Marshal(stockStatusResponse())
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
