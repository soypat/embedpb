package daemonpb

import (
	"testing"
	"time"

	stock "github.com/soypat/embedpb/internal/fixture/daemonstock"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/durationpb"
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

func TestRuntimeDaemonForwarding(t *testing.T) {
	stockMsg := &stock.ForwardingRulesResponse{Rules: []*stock.ForwardingRule{
		{
			Protocol:          "tcp",
			DestinationPort:   &stock.PortInfo{PortSelection: &stock.PortInfo_Port{Port: 8443}},
			TranslatedAddress: "100.64.0.50",
			TranslatedPort:    &stock.PortInfo{PortSelection: &stock.PortInfo_Port{Port: 443}},
		},
		{
			Protocol:        "udp",
			DestinationPort: &stock.PortInfo{PortSelection: &stock.PortInfo_Range_{Range: &stock.PortInfo_Range{Start: 5000, End: 6000}}},
			TranslatedPort:  &stock.PortInfo{PortSelection: &stock.PortInfo_Range_{Range: &stock.PortInfo_Range{Start: 15000, End: 16000}}},
		},
	}}
	b, err := proto.Marshal(stockMsg)
	if err != nil {
		t.Fatal(err)
	}
	var out ForwardingRulesResponse
	if err := proto.Unmarshal(b, &out); err != nil {
		t.Fatal(err)
	}
	if out.GetRules()[0].GetDestinationPort().GetPort() != 8443 ||
		out.GetRules()[1].GetDestinationPort().GetRange().GetEnd() != 6000 ||
		out.GetRules()[1].GetTranslatedPort().GetRange().GetStart() != 15000 {
		t.Fatalf("decoded runtime daemon forwarding wrong: %+v", &out)
	}
}

func TestRuntimeDaemonLogin(t *testing.T) {
	stockMsg := &stock.LoginRequest{
		SetupKey:         "setup-runtime",
		ManagementUrl:    "https://management.netbird.test",
		RosenpassEnabled: boolp(true),
		InterfaceName:    stringp("wg-runtime"),
		DnsRouteInterval: durationpb.New(45 * time.Second),
		DnsLabels:        []string{"runtime", "daemon"},
	}
	b, err := proto.Marshal(stockMsg)
	if err != nil {
		t.Fatal(err)
	}
	var out LoginRequest
	if err := proto.Unmarshal(b, &out); err != nil {
		t.Fatal(err)
	}
	if !out.GetRosenpassEnabled() ||
		out.GetInterfaceName() != "wg-runtime" ||
		out.GetDnsRouteInterval().AsDuration() != 45*time.Second ||
		out.GetDnsLabels()[1] != "daemon" {
		t.Fatalf("decoded runtime daemon login wrong: %+v", &out)
	}
}
