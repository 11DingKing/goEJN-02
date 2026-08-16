package http

import (
	"encoding/json"
	"net/http"
	"testing"
)

// TestDuplicateAcceptReturnsConflict checks the HTTP status a dispatcher sees
// when a second repair team posts an accept for a work order that is already
// taken: the API must answer with a 409 conflict.
func TestDuplicateAcceptReturnsConflict(t *testing.T) {
	h, _ := testHandler(t)
	routes := h.Routes()

	w := doRequest(t, routes, "POST", "/api/v1/anomalies", map[string]any{
		"entity_type": "diesel_generator", "entity_id": "dg-1", "description": "fuel leak",
	})
	if w.Code != http.StatusCreated {
		t.Fatalf("report anomaly: %d %s", w.Code, w.Body.String())
	}
	var wo map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &wo); err != nil {
		t.Fatalf("decode work order: %v", err)
	}
	woID, ok := wo["id"].(string)
	if !ok {
		t.Fatalf("work order id missing in %s", w.Body.String())
	}

	w = doRequest(t, routes, "POST", "/api/v1/workorders/"+woID+"/accept", map[string]any{"assignee": "repair-li"})
	if w.Code != http.StatusOK {
		t.Fatalf("first accept: %d %s", w.Code, w.Body.String())
	}

	w = doRequest(t, routes, "POST", "/api/v1/workorders/"+woID+"/accept", map[string]any{"assignee": "repair-wang"})
	if w.Code != http.StatusConflict {
		t.Fatalf("second accept by another team: status = %d, want 409; body = %s", w.Code, w.Body.String())
	}
}
