package managementjobpb

import (
	"testing"

	stock "github.com/soypat/embedpb/internal/fixture/managementjobstock"

	"google.golang.org/protobuf/proto"
)

func TestRuntimeJobResponseInterop(t *testing.T) {
	stockMsg := &stock.JobResponse{
		ID:     []byte("job-runtime"),
		Status: stock.JobStatus_failed,
		Reason: []byte("upload failed"),
		WorkloadResults: &stock.JobResponse_Bundle{
			Bundle: &stock.BundleResult{UploadKey: "runtime-upload-key"},
		},
	}
	b, err := proto.Marshal(stockMsg)
	if err != nil {
		t.Fatal(err)
	}
	var out JobResponse
	if err := proto.Unmarshal(b, &out); err != nil {
		t.Fatal(err)
	}
	if out.GetStatus() != JobStatus_failed ||
		string(out.GetReason()) != "upload failed" ||
		out.GetBundle().GetUploadKey() != "runtime-upload-key" {
		t.Fatalf("decoded runtime job response wrong: %+v", &out)
	}
}
