package main

import (
	"embed"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
)

//go:embed static/index.html static/dev.html
var staticFS embed.FS

type Config struct {
	Port     int
	ApiKey   string
	Provider string
	Model    string
	Dev      bool
	Spy      bool
}

func getEnvOrDefault(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

// resolveAPIKey returns a trimmed key from the flag value, falling back to FINAL_AI_API_KEY.
func resolveAPIKey(flagValue string) string {
	key := strings.TrimSpace(flagValue)
	if key == "" {
		key = strings.TrimSpace(os.Getenv("FINAL_AI_API_KEY"))
	}
	return key
}

func requireAPIKey(key string) {
	if key != "" {
		return
	}
	fmt.Fprintln(os.Stderr, "final-ai: API key required.")
	fmt.Fprintln(os.Stderr, "  Supply via:  -apikey theKeyText")
	fmt.Fprintln(os.Stderr, "  or set env:  FINAL_AI_API_KEY")
	os.Exit(1)
}

func main() {
	cfg := &Config{}
	defApiKey := getEnvOrDefault("FINAL_AI_API_KEY", "")

	flag.IntVar(&cfg.Port, "port", 8088, "port to listen on")
	flag.StringVar(&cfg.ApiKey, "apikey", defApiKey, "API key")
	flag.StringVar(&cfg.Provider, "provider", "grok", "AI provider (grok only in v0)")
	flag.StringVar(&cfg.Model, "model", "grok-3", "chat model id")
	flag.BoolVar(&cfg.Dev, "dev", false, "enable Dev tab (Final Session console)")
	flag.BoolVar(&cfg.Spy, "spy", false, "log Final tool req/res to stdout (temporary debug)")
	flag.Parse()

	cfg.ApiKey = resolveAPIKey(cfg.ApiKey)
	requireAPIKey(cfg.ApiKey)

	srv, err := newServer(cfg)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/api/chat", srv.handleChat)
	if cfg.Dev {
		mux.HandleFunc("/api/final/run", srv.handleFinalRun)
	}

	indexPage := "index.html"
	if cfg.Dev {
		indexPage = "dev.html"
	}
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		data, err := staticFS.ReadFile("static/" + indexPage)
		if err != nil {
			http.Error(w, "page not found", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write(data)
	})

	addr := fmt.Sprintf(":%d", cfg.Port)
	if cfg.Dev {
		log.Printf("final-ai listening on http://localhost%s (dev mode: AI Chat + Dev tabs)", addr)
	} else {
		log.Printf("final-ai listening on http://localhost%s", addr)
	}
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal(err)
	}
}
