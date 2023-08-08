package main

import (
	"database/sql"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestServer(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("opening database: %v", err)
	}
	defer db.Close()
	model, err := NewSQLModel(db)
	exitOnError(err)

	server, err := NewServer(ServerOpts{
		model:  model,
		logger: log.Default(),
	})
	if err != nil {
		t.Fatalf("Error creating server")
	}

	truthContent := "foo"

	// Create a drop
	body := strings.NewReader(truthContent)
	r := httptest.NewRequest(http.MethodPost, "/d", body)
	recorder := httptest.NewRecorder()
	server.ServeHTTP(recorder, r)
	ensureCode(t, recorder, http.StatusOK)
	// Get the url we generate
	rBodyUrl := recorder.Body.String()
	u, err := url.Parse(rBodyUrl)
	if err != nil {
		log.Fatal(err)
	}

	// The payload should be what we sent to the drop
	r = httptest.NewRequest(http.MethodGet, u.Path, nil)
	recorder = httptest.NewRecorder()
	server.ServeHTTP(recorder, r)
	ensureCode(t, recorder, http.StatusOK)
	// Get the url we generate
	rBody := recorder.Body.String()
	ensureString(t, truthContent, rBody)

	// The drop should not exist anymore
	r = httptest.NewRequest(http.MethodGet, u.Path, nil)
	recorder = httptest.NewRecorder()
	server.ServeHTTP(recorder, r)
	ensureCode(t, recorder, http.StatusNotFound)
}

func ensureString(t *testing.T, got, want string) {
	t.Helper()
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func ensureInt(t *testing.T, got, want int) {
	t.Helper()
	if got != want {
		t.Fatalf("got %d, want %d", got, want)
	}
}

func ensureCode(t *testing.T, recorder *httptest.ResponseRecorder, expected int) {
	t.Helper()
	if recorder.Code != expected {
		t.Fatalf("got code %d, want %d, response body:\n%s",
			recorder.Code, expected, recorder.Body.String())
	}
}
