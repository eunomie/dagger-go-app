package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestValidateScoreInput(t *testing.T) {
	if err := validateScoreInput("", 0); err == nil {
		t.Errorf("expected error for empty name")
	}
	long := make([]byte, 51)
	for i := range long { long[i] = 'a' }
	if err := validateScoreInput(string(long), 0); err == nil {
		t.Errorf("expected error for too long name")
	}
	if err := validateScoreInput("ok", -1); err == nil {
		t.Errorf("expected error for negative score")
	}
	if err := validateScoreInput("ok", 0); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestHandleHealth(t *testing.T) {
	s := &Server{}
	req := httptest.NewRequest(http.MethodGet, "/api/healthz", nil)
	rec := httptest.NewRecorder()
	s.handleHealth(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var body map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	if body["status"] != "ok" {
		t.Fatalf("unexpected body: %v", body)
	}
}

func TestGetScores(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil { t.Fatalf("mock err: %v", err) }
	defer db.Close()
	s := &Server{DB: db}

	rows := sqlmock.NewRows([]string{"id", "name", "score", "created_at"}).
		AddRow(1, "Alice", 200, time.Now()).
		AddRow(2, "Bob", 100, time.Now().Add(-time.Hour))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, name, score, created_at FROM scores ORDER BY score DESC, created_at ASC LIMIT $1")).
		WithArgs(10).WillReturnRows(rows)

	req := httptest.NewRequest(http.MethodGet, "/api/scores?limit=10", nil)
	rec := httptest.NewRecorder()
	s.getScores(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var body struct { Scores []Score `json:"scores"` }
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	if len(body.Scores) != 2 {
		t.Fatalf("expected 2 scores, got %d", len(body.Scores))
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestPostScore_Validation(t *testing.T) {
	s := &Server{DB: &sql.DB{}} // DB won't be used on validation failures
	cases := []struct{
		name string
		contentType string
		body string
		expected int
	}{
		{"bad-ct", "text/plain", `{}`, http.StatusUnsupportedMediaType},
		{"bad-json", "application/json", `{`, http.StatusBadRequest},
		{"empty-name", "application/json", `{"name":" ","score":1}`, http.StatusBadRequest},
		{"too-long", "application/json", `{"name":"` + string(bytes.Repeat([]byte{'a'}, 51)) + `","score":1}`, http.StatusBadRequest},
		{"negative", "application/json", `{"name":"ok","score":-1}`, http.StatusBadRequest},
	}
	for _, tc := range cases {
		req := httptest.NewRequest(http.MethodPost, "/api/scores", bytes.NewBufferString(tc.body))
		req.Header.Set("Content-Type", tc.contentType)
		rec := httptest.NewRecorder()
		s.postScore(rec, req)
		if rec.Code != tc.expected {
			t.Errorf("%s: expected %d, got %d", tc.name, tc.expected, rec.Code)
		}
	}
}

func TestPostScore_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil { t.Fatalf("mock err: %v", err) }
	defer db.Close()
	s := &Server{DB: db}
	now := time.Now()
	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO scores(name, score) VALUES($1, $2) RETURNING id, name, score, created_at")).
		WithArgs("Alice", 123).
		WillReturnRows(sqlmock.NewRows([]string{"id","name","score","created_at"}).AddRow(1, "Alice", 123, now))

	body := map[string]any{"name":"Alice","score":123}
	b, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, "/api/scores", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	s.postScore(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", rec.Code)
	}
	var out Score
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	if out.ID != 1 || out.Name != "Alice" || out.Score != 123 {
		t.Fatalf("unexpected body: %+v", out)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}
