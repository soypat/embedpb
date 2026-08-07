package managementnetpb

import (
	"testing"

	stock "github.com/soypat/embedpb/internal/fixture/managementnetstock"

	"google.golang.org/protobuf/proto"
)

func TestRuntimeManagementNet(t *testing.T) {
	stockMsg := &stock.SyncResponse{
		NetbirdConfig: &stock.NetbirdConfig{
			Relay: &stock.RelayConfig{Urls: []string{"rels://relay-runtime.netbird.test"}},
		},
		NetworkMap: &stock.NetworkMap{
			Serial: 7,
			Routes: []*stock.Route{{
				ID:      "route-runtime",
				Network: "172.16.0.0/16",
				Domains: []string{"runtime.example.test"},
			}},
			ForwardingRules: []*stock.ForwardingRule{{
				Protocol:          stock.RuleProtocol_TCP,
				DestinationPort:   &stock.PortInfo{PortSelection: &stock.PortInfo_Port{Port: 8443}},
				TranslatedAddress: []byte{100, 64, 0, 50},
				TranslatedPort:    &stock.PortInfo{PortSelection: &stock.PortInfo_Port{Port: 443}},
			}},
		},
	}
	b, err := proto.Marshal(stockMsg)
	if err != nil {
		t.Fatal(err)
	}
	var out SyncResponse
	if err := proto.Unmarshal(b, &out); err != nil {
		t.Fatal(err)
	}
	if out.GetNetbirdConfig().GetRelay().GetUrls()[0] != "rels://relay-runtime.netbird.test" ||
		out.GetNetworkMap().GetRoutes()[0].GetID() != "route-runtime" ||
		out.GetNetworkMap().GetForwardingRules()[0].GetDestinationPort().GetPort() != 8443 {
		t.Fatalf("decoded runtime management net wrong: %+v", &out)
	}
}
