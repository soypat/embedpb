package managementjobpb

import (
	"bytes"
	"encoding/json"
	"reflect"
	"testing"

	stock "github.com/soypat/embedpb/internal/fixture/managementjobstock"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

func sampleRequest() *JobRequest {
	return &JobRequest{
		ID: []byte("job-1"),
		WorkloadParameters: &JobRequest_Bundle{
			Bundle: &BundleParameters{
				BundleFor:     true,
				BundleForTime: 3600,
				LogFileCount:  5,
				Anonymize:     true,
			},
		},
	}
}

func stockRequest() *stock.JobRequest {
	return &stock.JobRequest{
		ID: []byte("job-1"),
		WorkloadParameters: &stock.JobRequest_Bundle{
			Bundle: &stock.BundleParameters{
				BundleFor:     true,
				BundleForTime: 3600,
				LogFileCount:  5,
				Anonymize:     true,
			},
		},
	}
}

func sampleResponse() *JobResponse {
	return &JobResponse{
		ID:     []byte("job-1"),
		Status: JobStatus_succeeded,
		Reason: []byte("done"),
		WorkloadResults: &JobResponse_Bundle{
			Bundle: &BundleResult{UploadKey: "bundle-upload-key"},
		},
	}
}

func stockResponse() *stock.JobResponse {
	return &stock.JobResponse{
		ID:     []byte("job-1"),
		Status: stock.JobStatus_succeeded,
		Reason: []byte("done"),
		WorkloadResults: &stock.JobResponse_Bundle{
			Bundle: &stock.BundleResult{UploadKey: "bundle-upload-key"},
		},
	}
}

func TestJobRequestInterop(t *testing.T) {
	ours, err := proto.Marshal(sampleRequest())
	if err != nil {
		t.Fatal(err)
	}
	var st stock.JobRequest
	if err := proto.Unmarshal(ours, &st); err != nil {
		t.Fatalf("stock unmarshal: %v", err)
	}
	if !bytes.Equal(st.GetID(), []byte("job-1")) ||
		!st.GetBundle().GetBundleFor() ||
		st.GetBundle().GetLogFileCount() != 5 {
		t.Fatalf("stock decoded job request wrong: %+v", &st)
	}

	stockBytes, err := proto.Marshal(stockRequest())
	if err != nil {
		t.Fatal(err)
	}
	var mine JobRequest
	if err := proto.Unmarshal(stockBytes, &mine); err != nil {
		t.Fatalf("our unmarshal: %v", err)
	}
	if !bytes.Equal(mine.GetID(), []byte("job-1")) ||
		mine.GetBundle().GetBundleForTime() != 3600 ||
		!mine.GetBundle().GetAnonymize() {
		t.Fatalf("we decoded stock job request wrong: %+v", &mine)
	}
}

func TestJobResponseInterop(t *testing.T) {
	ours, err := proto.Marshal(sampleResponse())
	if err != nil {
		t.Fatal(err)
	}
	var st stock.JobResponse
	if err := proto.Unmarshal(ours, &st); err != nil {
		t.Fatalf("stock unmarshal: %v", err)
	}
	if st.GetStatus() != stock.JobStatus_succeeded ||
		string(st.GetReason()) != "done" ||
		st.GetBundle().GetUploadKey() != "bundle-upload-key" {
		t.Fatalf("stock decoded job response wrong: %+v", &st)
	}

	stockBytes, err := proto.Marshal(stockResponse())
	if err != nil {
		t.Fatal(err)
	}
	var mine JobResponse
	if err := proto.Unmarshal(stockBytes, &mine); err != nil {
		t.Fatalf("our unmarshal: %v", err)
	}
	if mine.GetStatus() != JobStatus_succeeded ||
		string(mine.GetReason()) != "done" ||
		mine.GetBundle().GetUploadKey() != "bundle-upload-key" {
		t.Fatalf("we decoded stock job response wrong: %+v", &mine)
	}
}

func TestJobJSONMatchesProtojson(t *testing.T) {
	ours, err := sampleResponse().AppendJSON(nil)
	if err != nil {
		t.Fatal(err)
	}
	theirs, err := protojson.MarshalOptions{
		EmitUnpopulated: true,
		UseProtoNames:   true,
		AllowPartial:    true,
	}.Marshal(stockResponse())
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
