package managementnetpb

import (
	"encoding/json"
	"reflect"
	"testing"
	"time"

	stock "github.com/soypat/embedpb/internal/fixture/managementnetstock"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func sampleSyncResponse() *SyncResponse {
	return &SyncResponse{
		NetbirdConfig: &NetbirdConfig{
			Stuns: []*HostConfig{{Uri: "stun:stun.netbird.test:3478", Protocol: HostConfig_UDP}},
			Turns: []*ProtectedHostConfig{{
				HostConfig: &HostConfig{Uri: "turns:turn.netbird.test:443", Protocol: HostConfig_DTLS},
				User:       "turn-user",
				Password:   "turn-password",
			}},
			Signal: &HostConfig{Uri: "https://signal.netbird.test", Protocol: HostConfig_HTTPS},
			Relay: &RelayConfig{
				Urls:           []string{"rels://relay-a.netbird.test", "rels://relay-b.netbird.test"},
				TokenPayload:   "relay-payload",
				TokenSignature: "relay-signature",
			},
			Flow: &FlowConfig{
				Url:                "https://flow.netbird.test",
				TokenPayload:       "flow-payload",
				TokenSignature:     "flow-signature",
				Interval:           durationpb.New(90 * time.Second),
				Enabled:            true,
				Counters:           true,
				ExitNodeCollection: true,
				DnsCollection:      true,
			},
		},
		NetworkMap: &NetworkMap{
			Serial: 42,
			PeerConfig: &PeerConfig{
				Address:                         "100.64.0.1/32",
				Dns:                             "100.64.0.2",
				SshConfig:                       sshConfig(),
				Fqdn:                            "self.netbird.test",
				RoutingPeerDnsResolutionEnabled: true,
				LazyConnectionEnabled:           true,
				Mtu:                             1280,
				AutoUpdate:                      &AutoUpdateSettings{Version: "1.2.3", AlwaysUpdate: true},
				AddressV6:                       []byte{0xfd, 0, 0, 0, 0, 0, 0, 1, 64},
			},
			RemotePeers: []*RemotePeerConfig{{
				WgPubKey:     "peer-key",
				AllowedIps:   []string{"100.64.0.10/32", "fd00::10/128"},
				SshConfig:    sshConfig(),
				Fqdn:         "peer.netbird.test",
				AgentVersion: "0.99.0",
			}},
			Routes: []*Route{{
				ID:            "route-1",
				Network:       "10.10.0.0/16",
				NetworkType:   1,
				Peer:          "peer-id",
				Metric:        100,
				Masquerade:    true,
				NetID:         "net-1",
				Domains:       []string{"corp.example.test"},
				KeepRoute:     true,
				SkipAutoApply: true,
			}},
			DNSConfig: &DNSConfig{
				ServiceEnable: true,
				NameServerGroups: []*NameServerGroup{{
					NameServers:          []*NameServer{{IP: "1.1.1.1", NSType: 1, Port: 53}},
					Primary:              true,
					Domains:              []string{"example.test"},
					SearchDomainsEnabled: true,
				}},
				CustomZones: []*CustomZone{{
					Domain:               "corp.example.test",
					Records:              []*SimpleRecord{{Name: "app", Type: 1, Class: "IN", TTL: 60, RData: "100.64.0.20"}},
					SearchDomainDisabled: true,
					NonAuthoritative:     true,
				}},
				ForwarderPort: 5353,
			},
			OfflinePeers: []*RemotePeerConfig{{WgPubKey: "offline-peer", AllowedIps: []string{"100.64.0.11/32"}}},
			FirewallRules: []*FirewallRule{{
				PeerIP:         "100.64.0.10",
				Direction:      RuleDirection_IN,
				Action:         RuleAction_ACCEPT,
				Protocol:       RuleProtocol_CUSTOM,
				Port:           "443",
				PortInfo:       &PortInfo{PortSelection: &PortInfo_Port{Port: 443}},
				PolicyID:       []byte("policy-1"),
				CustomProtocol: 132,
				SourcePrefixes: [][]byte{{10, 0, 0, 0, 8}},
			}},
			RoutesFirewallRules: []*RouteFirewallRule{{
				SourceRanges:   []string{"100.64.0.0/10"},
				Action:         RuleAction_DROP,
				Destination:    "10.20.0.0/16",
				Protocol:       RuleProtocol_TCP,
				PortInfo:       &PortInfo{PortSelection: &PortInfo_Range_{Range: &PortInfo_Range{Start: 5000, End: 6000}}},
				IsDynamic:      true,
				Domains:        []string{"db.example.test"},
				CustomProtocol: 250,
				PolicyID:       []byte("policy-route"),
				RouteID:        "route-1",
			}},
			ForwardingRules: []*ForwardingRule{{
				Protocol:          RuleProtocol_UDP,
				DestinationPort:   &PortInfo{PortSelection: &PortInfo_Port{Port: 5353}},
				TranslatedAddress: []byte{100, 64, 0, 10},
				TranslatedPort:    &PortInfo{PortSelection: &PortInfo_Range_{Range: &PortInfo_Range{Start: 53, End: 54}}},
			}},
			SshAuth: &SSHAuth{
				UserIDClaim:     "sub",
				AuthorizedUsers: [][]byte{[]byte("hash-a"), []byte("hash-b")},
				MachineUsers:    map[string]*MachineUserIndexes{"root": {Indexes: []uint32{0}}, "admin": {Indexes: []uint32{0, 1}}},
			},
		},
		SessionExpiresAt: timestamppb.New(time.Unix(1800000000, 0).UTC()),
	}
}

func sshConfig() *SSHConfig {
	return &SSHConfig{
		SshEnabled: true,
		SshPubKey:  []byte("ssh-ed25519 AAAAMOCK"),
		JwtConfig: &JWTConfig{
			Issuer:       "https://idp.netbird.test",
			Audience:     "netbird",
			KeysLocation: "https://idp.netbird.test/keys",
			MaxTokenAge:  3600,
			Audiences:    []string{"netbird", "netbird-alt"},
		},
	}
}

func stockSyncResponse() *stock.SyncResponse {
	b, err := proto.Marshal(sampleSyncResponse())
	if err != nil {
		panic(err)
	}
	var out stock.SyncResponse
	if err := proto.Unmarshal(b, &out); err != nil {
		panic(err)
	}
	return &out
}

func TestManagementNetInterop(t *testing.T) {
	ours, err := proto.Marshal(sampleSyncResponse())
	if err != nil {
		t.Fatal(err)
	}
	var st stock.SyncResponse
	if err := proto.Unmarshal(ours, &st); err != nil {
		t.Fatalf("stock unmarshal: %v", err)
	}
	if st.GetNetbirdConfig().GetRelay().GetUrls()[1] != "rels://relay-b.netbird.test" ||
		st.GetNetworkMap().GetRoutes()[0].GetDomains()[0] != "corp.example.test" ||
		st.GetNetworkMap().GetForwardingRules()[0].GetTranslatedPort().GetRange().GetEnd() != 54 {
		t.Fatalf("stock decoded management net wrong: %+v", &st)
	}

	stockBytes, err := proto.Marshal(stockSyncResponse())
	if err != nil {
		t.Fatal(err)
	}
	var mine SyncResponse
	if err := proto.Unmarshal(stockBytes, &mine); err != nil {
		t.Fatalf("our unmarshal: %v", err)
	}
	if mine.GetNetworkMap().GetDNSConfig().GetCustomZones()[0].GetRecords()[0].GetRData() != "100.64.0.20" ||
		mine.GetNetworkMap().GetRoutesFirewallRules()[0].GetPortInfo().GetRange().GetStart() != 5000 ||
		mine.GetNetworkMap().GetSshAuth().GetMachineUsers()["admin"].GetIndexes()[1] != 1 {
		t.Fatalf("we decoded stock management net wrong: %+v", &mine)
	}
}

func TestManagementNetJSONMatchesProtojson(t *testing.T) {
	ours, err := sampleSyncResponse().AppendJSON(nil)
	if err != nil {
		t.Fatal(err)
	}
	theirs, err := protojson.MarshalOptions{
		EmitUnpopulated: true,
		UseProtoNames:   true,
		AllowPartial:    true,
	}.Marshal(stockSyncResponse())
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
