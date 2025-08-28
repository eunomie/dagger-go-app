package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	_ "github.com/lib/pq"
)

type Server struct {
	DB *sql.DB
}

type Score struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Score     int       `json:"score"`
	CreatedAt time.Time `json:"created_at"`
}

func main() {
	addr := getEnv("ADDR", ":8080")

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Println("WARNING: DATABASE_URL is not set. The server will not start without a PostgreSQL connection as scores must be persisted.")
		log.Fatal("Please set DATABASE_URL, e.g. postgres://user:pass@localhost:5432/dbname?sslmode=disable")
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()
	if err := pingWithRetry(db, 5, time.Second); err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	if err := ensureSchema(db); err != nil {
		log.Fatalf("failed to ensure schema: %v", err)
	}

	s := &Server{DB: db}

	mux := http.NewServeMux()
	// API routes
	mux.HandleFunc("/api/healthz", s.handleHealth)
	mux.HandleFunc("/api/scores", s.handleScores)

	// Static files (built frontend)
	static := http.FileServer(http.Dir("web/dist"))
	mux.Handle("/", spaHandler(static, "web/dist/index.html"))

	log.Printf("Server listening on %s", addr)
	if err := http.ListenAndServe(addr, withLogging(cors(mux))); err != nil {
		log.Fatal(err)
	}
}

func getEnv(key, def string) string { if v := os.Getenv(key); v != "" { return v }; return def }

func pingWithRetry(db *sql.DB, attempts int, delay time.Duration) error {
	var err error
	for i := 0; i < attempts; i++ {
		err = db.Ping()
		if err == nil {
			return nil
		}
		log.Printf("db ping failed (attempt %d/%d): %v", i+1, attempts, err)
		time.Sleep(delay)
	}
	return err
}

func ensureSchema(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS scores (
			id SERIAL PRIMARY KEY,
			name TEXT NOT NULL,
			score INTEGER NOT NULL CHECK (score >= 0),
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);
		CREATE INDEX IF NOT EXISTS idx_scores_score ON scores(score DESC);
	`)
	return err
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func (s *Server) handleScores(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.getScores(w, r)
	case http.MethodPost:
		s.postScore(w, r)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (s *Server) getScores(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	limit := 10
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 100 {
			limit = n
		}
	}
	rows, err := s.DB.Query(`SELECT id, name, score, created_at FROM scores ORDER BY score DESC, created_at ASC LIMIT $1`, limit)
	if err != nil {
		httpError(w, http.StatusInternalServerError, "failed to query scores")
		return
	}
	defer rows.Close()
	var scores []Score
	for rows.Next() {
		var sc Score
		if err := rows.Scan(&sc.ID, &sc.Name, &sc.Score, &sc.CreatedAt); err != nil {
			httpError(w, http.StatusInternalServerError, "failed to scan score")
			return
		}
		scores = append(scores, sc)
	}
	_ = json.NewEncoder(w).Encode(map[string]any{"scores": scores})
}

func (s *Server) postScore(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if ct := r.Header.Get("Content-Type"); ct != "" && !strings.Contains(ct, "application/json") {
		httpError(w, http.StatusUnsupportedMediaType, "Content-Type must be application/json")
		return
	}
	var in struct {
		Name  string `json:"name"`
		Score int    `json:"score"`
	}
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		httpError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	in.Name = strings.TrimSpace(in.Name)
	if err := validateScoreInput(in.Name, in.Score); err != nil {
		httpError(w, http.StatusBadRequest, err.Error())
		return
	}
	var out Score
	err := s.DB.QueryRow(`INSERT INTO scores(name, score) VALUES($1, $2) RETURNING id, name, score, created_at`, in.Name, in.Score).
		Scan(&out.ID, &out.Name, &out.Score, &out.CreatedAt)
	if err != nil {
		httpError(w, http.StatusInternalServerError, "failed to insert score")
		return
	}
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(out)
}

func validateScoreInput(name string, score int) error {
	if name == "" {
		return errors.New("name is required")
	}
	if len(name) > 50 {
		return errors.New("name too long (max 50)")
	}
	if score < 0 {
		return errors.New("score must be non-negative")
	}
	return nil
}

// spaHandler tries to serve static files, and falls back to index.html for SPA routes.
func spaHandler(fs http.Handler, indexPath string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Only GET/HEAD are safe to serve as files
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			fs.ServeHTTP(w, r)
			return
		}
		p := r.URL.Path
		if p == "/" || p == "" {
			http.ServeFile(w, r, indexPath)
			return
		}
		clean := path.Clean(p)
		full := filepath.Join("web/dist", clean)
		if info, err := os.Stat(full); err == nil && !info.IsDir() {
			fs.ServeHTTP(w, r)
			return
		}
		// Fallback to index.html for client-side routing
		http.ServeFile(w, r, indexPath)
	})
}

func withLogging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s %s", r.Method, r.URL.Path, time.Since(start))
	})
}

func cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin != "" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Vary", "Origin")
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
			w.Header().Set("Access-Control-Allow-Methods", "GET,POST,OPTIONS")
		}
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func httpError(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": msg})
}
