package daemonpb

import (
	"testing"
	"time"

	stock "github.com/soypat/embedpb/internal/fixture/daemonstock"
	"github.com/soypat/embedpb/internal/fixture/fixtest"

	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// sampleLoginRequest is the daemon's proto3-optional workhorse: 32 optional
// fields, a deprecated one, and a Duration.
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
		RosenpassEnabled:              new(true),
		InterfaceName:                 new("wg0"),
		WireguardPort:                 new(int64(51820)),
		OptionalPreSharedKey:          new("modern-key"),
		DisableAutoConnect:            new(true),
		ServerSSHAllowed:              new(true),
		RosenpassPermissive:           new(false),
		ExtraIFaceBlacklist:           []string{"docker0", "virbr0"},
		NetworkMonitor:                new(true),
		DnsRouteInterval:              durationpb.New(30 * time.Second),
		DisableClientRoutes:           new(false),
		DisableServerRoutes:           new(true),
		DisableDns:                    new(false),
		DisableFirewall:               new(true),
		BlockLanAccess:                new(true),
		DisableNotifications:          new(false),
		DnsLabels:                     []string{"prod", "ssh"},
		CleanDNSLabels:                true,
		LazyConnectionEnabled:         new(true),
		BlockInbound:                  new(true),
		ProfileName:                   new("work"),
		Username:                      new("user@example.test"),
		Mtu:                           new(int64(1280)),
		Hint:                          new("user@example.test"),
		EnableSSHRoot:                 new(true),
		EnableSSHSFTP:                 new(true),
		EnableSSHLocalPortForwarding:  new(true),
		EnableSSHRemotePortForwarding: new(false),
		DisableSSHAuth:                new(false),
		SshJWTCacheTTL:                new(int32(600)),
		DisableIpv6:                   new(false),
	}
}

// sampleStatusResponse carries the nested enums (SystemEvent.Severity /
// .Category), Timestamps and a map<string,string>.
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

// sampleForwardingRules exercises both arms of the nested PortInfo.Range oneof.
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

// sampleNetwork carries map<string,IPList> — a map with a message value.
func sampleNetwork() *Network {
	return &Network{
		ID:       "network-a",
		Range:    "10.10.0.0/16",
		Selected: true,
		Domains:  []string{"corp.example.test"},
		ResolvedIPs: map[string]*IPList{
			"app.corp.example.test": {Ips: []string{"100.64.0.10", "fd00::10"}},
		},
	}
}

// sampleExposeServiceEvent is a single-arm message oneof.
func sampleExposeServiceEvent() *ExposeServiceEvent {
	return &ExposeServiceEvent{Event: &ExposeServiceEvent_Ready{
		Ready: &ExposeServiceReady{ServiceName: "app"},
	}}
}

func TestDaemon(t *testing.T) {
	fixtest.Run(t, []fixtest.Case{
		{Name: "login-request", Ours: sampleLoginRequest(), Stock: new(stock.LoginRequest)},
		{Name: "login-request-optionals-unset", Ours: &LoginRequest{Hostname: "peer-a"}, Stock: new(stock.LoginRequest)},
		{Name: "status-response", Ours: sampleStatusResponse(), Stock: new(stock.StatusResponse)},
		{Name: "forwarding-rules", Ours: sampleForwardingRules(), Stock: new(stock.ForwardingRulesResponse)},
		{Name: "network", Ours: sampleNetwork(), Stock: new(stock.Network)},
		{Name: "expose-service-event", Ours: sampleExposeServiceEvent(), Stock: new(stock.ExposeServiceEvent)},
	})
}
