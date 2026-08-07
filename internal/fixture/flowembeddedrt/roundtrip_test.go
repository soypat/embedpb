package flowpb

import (
	"testing"
	"time"

	stock "github.com/soypat/embedpb/internal/fixture/flowstock"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestRuntimeFlowEventInterop(t *testing.T) {
	stockMsg := &stock.FlowEvent{
		EventId:     []byte("event-runtime"),
		Timestamp:   timestamppb.New(time.Unix(1800000002, 789000000).UTC()),
		IsInitiator: true,
		FlowFields: &stock.FlowFields{
			FlowId:         []byte("flow-runtime"),
			Type:           stock.Type_TYPE_ALLOWED,
			Direction:      stock.Direction_DIRECTION_EGRESS,
			Protocol:       17,
			ConnectionInfo: &stock.FlowFields_PortInfo{PortInfo: &stock.PortInfo{SourcePort: 5353, DestPort: 53}},
			TxPackets:      4,
			TxBytes:        512,
		},
	}
	b, err := proto.Marshal(stockMsg)
	if err != nil {
		t.Fatal(err)
	}
	var out FlowEvent
	if err := proto.Unmarshal(b, &out); err != nil {
		t.Fatal(err)
	}
	if out.GetFlowFields().GetPortInfo().GetSourcePort() != 5353 ||
		out.GetFlowFields().GetProtocol() != 17 ||
		out.GetTimestamp().GetNanos() != 789000000 {
		t.Fatalf("decoded runtime flow wrong: %+v", &out)
	}
}
