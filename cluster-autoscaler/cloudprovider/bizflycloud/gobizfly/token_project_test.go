package gobizfly

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
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
