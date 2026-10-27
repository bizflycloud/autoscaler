package gobizfly

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestTokenCreateSendsProjectID(t *testing.T) {
	setup()
	defer teardown()
	client.projectID = "project-1"

	var got TokenCreateRequest
	mux.HandleFunc("/api/token", func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
			t.Errorf("decode request: %v", err)
		}
		fmt.Fprint(w, `{"token":"tok","project_id":"project-1"}`)
	})
	mux.HandleFunc("/api/auth/service", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"services":[]}`)
	})

	_, err := client.Token.Create(context.Background(), &TokenCreateRequest{
		AuthMethod:    "application_credential",
		AppCredID:     "id",
		AppCredSecret: "secret",
	})
	if err != nil {
		t.Fatalf("create token: %v", err)
	}
	if got.ProjectID != "project-1" {
		t.Fatalf("project_id = %q, want project-1", got.ProjectID)
	}
}

// A new client has no service catalog yet, so the token request must go to
// the configured API URL. It used to go to the production default, and a
// staging credential was rejected there with "credential không đúng".
func TestTokenCreateUsesConfiguredAPIURL(t *testing.T) {
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	defer server.Close()

	var tokenRequests int
	mux.HandleFunc("/api/token", func(w http.ResponseWriter, r *http.Request) {
		tokenRequests++
		fmt.Fprint(w, `{"token":"tok","project_id":"project-1"}`)
	})
	mux.HandleFunc("/api/auth/service", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"services":[]}`)
	})

	c, err := NewClient(WithAPIUrl(server.URL), WithRegionName("HN"))
	if err != nil {
		t.Fatalf("new client: %v", err)
	}
	if _, err := c.Token.Create(context.Background(), &TokenCreateRequest{
		AuthMethod:    "application_credential",
		AppCredID:     "id",
		AppCredSecret: "secret",
	}); err != nil {
		t.Fatalf("create token: %v", err)
	}
	if tokenRequests != 1 {
		t.Fatalf("token requests to the configured API = %d, want 1", tokenRequests)
	}
}
