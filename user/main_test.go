package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestLoginHandler(t *testing.T) {
	reqBody := `{"user_id": "driver-123"}`
	req, err := http.NewRequest(http.MethodPost, "/login", bytes.NewBufferString(reqBody))
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(loginHandler)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	var response map[string]string
	if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if val, ok := response["token"]; !ok || len(val) == 0 {
		t.Error("expected valid 'token' field in response")
	}
}
