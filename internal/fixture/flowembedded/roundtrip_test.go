package flowpb

import (
	"encoding/json"
	"reflect"
	"testing"
	"time"

	stock "github.com/soypat/embedpb/internal/fixture/flowstock"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func samplePortFlowEvent() *FlowEvent {
	return &FlowEvent{
		EventId:     []byte("event-1"),
		Timestamp:   timestamppb.New(time.Unix(1800000000, 456000000).UTC()),
		PublicKey:   []byte("public-key"),
		IsInitiator: true,
		FlowFields: &FlowFields{
			FlowId:           []byte("flow-1"),
			Type:             Type_TYPE_ALLOWED,
			RuleId:           []byte("rule-1"),
			Direction:        Direction_DIRECTION_EGRESS,
			Protocol:         6,
			SourceIp:         []byte{10, 0, 0, 1},
			DestIp:           []byte{10, 0, 0, 2},
			ConnectionInfo:   &FlowFields_PortInfo{PortInfo: &PortInfo{SourcePort: 12345, DestPort: 443}},
			RxPackets:        10,
			TxPackets:        20,
			RxBytes:          1000,
			TxBytes:          2000,
			SourceResourceId: []byte("src-resource"),
			DestResourceId:   []byte("dst-resource"),
		},
	}
}

func stockPortFlowEvent() *stock.FlowEvent {
	return &stock.FlowEvent{
		EventId:     []byte("event-1"),
		Timestamp:   timestamppb.New(time.Unix(1800000000, 456000000).UTC()),
		PublicKey:   []byte("public-key"),
		IsInitiator: true,
		FlowFields: &stock.FlowFields{
			FlowId:           []byte("flow-1"),
			Type:             stock.Type_TYPE_ALLOWED,
			RuleId:           []byte("rule-1"),
			Direction:        stock.Direction_DIRECTION_EGRESS,
			Protocol:         6,
			SourceIp:         []byte{10, 0, 0, 1},
			DestIp:           []byte{10, 0, 0, 2},
			ConnectionInfo:   &stock.FlowFields_PortInfo{PortInfo: &stock.PortInfo{SourcePort: 12345, DestPort: 443}},
			RxPackets:        10,
			TxPackets:        20,
			RxBytes:          1000,
			TxBytes:          2000,
			SourceResourceId: []byte("src-resource"),
			DestResourceId:   []byte("dst-resource"),
		},
	}
}

func stockICMPFlowEvent() *stock.FlowEvent {
	return &stock.FlowEvent{
		EventId:   []byte("event-icmp"),
		Timestamp: timestamppb.New(time.Unix(1800000001, 0).UTC()),
		FlowFields: &stock.FlowFields{
			FlowId:         []byte("flow-icmp"),
			Type:           stock.Type_TYPE_DENIED,
			Direction:      stock.Direction_DIRECTION_INGRESS,
			Protocol:       1,
			ConnectionInfo: &stock.FlowFields_IcmpInfo{IcmpInfo: &stock.ICMPInfo{IcmpType: 8, IcmpCode: 0}},
			RxPackets:      3,
			RxBytes:        128,
		},
	}
}

func TestFlowEventInterop(t *testing.T) {
	ours, err := proto.Marshal(samplePortFlowEvent())
	if err != nil {
		t.Fatal(err)
	}
	var st stock.FlowEvent
	if err := proto.Unmarshal(ours, &st); err != nil {
		t.Fatalf("stock unmarshal: %v", err)
	}
	if st.GetFlowFields().GetPortInfo().GetDestPort() != 443 ||
		st.GetFlowFields().GetTxBytes() != 2000 ||
		!st.GetIsInitiator() {
		t.Fatalf("stock decoded flow wrong: %+v", &st)
	}

	stockBytes, err := proto.Marshal(stockICMPFlowEvent())
	if err != nil {
		t.Fatal(err)
	}
	var mine FlowEvent
	if err := proto.Unmarshal(stockBytes, &mine); err != nil {
		t.Fatalf("our unmarshal: %v", err)
	}
	if mine.GetFlowFields().GetIcmpInfo().GetIcmpType() != 8 ||
		mine.GetFlowFields().GetRxBytes() != 128 ||
		mine.GetFlowFields().GetType() != Type_TYPE_DENIED {
		t.Fatalf("we decoded stock flow wrong: %+v", &mine)
	}
}

func TestFlowJSONMatchesProtojson(t *testing.T) {
	ours, err := samplePortFlowEvent().AppendJSON(nil)
	if err != nil {
		t.Fatal(err)
	}
	theirs, err := protojson.MarshalOptions{
		EmitUnpopulated: true,
		UseProtoNames:   true,
		AllowPartial:    true,
	}.Marshal(stockPortFlowEvent())
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

func TestFlowEventAck(t *testing.T) {
	ack := &FlowEventAck{EventId: []byte("event-1"), IsInitiator: true}
	b, err := proto.Marshal(ack)
	if err != nil {
		t.Fatal(err)
	}
	var out FlowEventAck
	if err := proto.Unmarshal(b, &out); err != nil {
		t.Fatal(err)
	}
	if string(out.GetEventId()) != "event-1" || !out.GetIsInitiator() {
		t.Fatalf("decoded ack wrong: %+v", &out)
	}
}
