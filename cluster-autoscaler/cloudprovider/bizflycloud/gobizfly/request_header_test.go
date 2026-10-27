package gobizfly

import (
	"context"
	"net/http"
	"testing"
)

func TestNewRequestSendsProjectIDHeader(t *testing.T) {
	setup()
	defer teardown()
	client.projectID = "project-1"

	req, err := client.NewRequest(context.Background(), http.MethodGet, authServiceName, "/token", nil)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	if got := req.Header.Get("X-Project-ID"); got != "project-1" {
		t.Fatalf("X-Project-ID = %q, want project-1", got)
	}
	if got := req.Header.Get("X-Tenant-Id"); got != "project-1" {
		t.Fatalf("X-Tenant-Id = %q, want project-1", got)
	}
}
