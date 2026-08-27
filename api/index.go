package handler

import (
	"crypto/sha256"
	"embed"
	"encoding/hex"
	"html/template"
	"net/http"
	"strings"
	"sync"
)

//go:embed views/*.html
var templateFiles embed.FS

type URLStore struct {
	sync.RWMutex
	data map[string]string
}

var store = URLStore{
	data: make(map[string]string),
}

func Handler(w http.ResponseWriter, r *http.Request) {

	path := r.URL.Path

	// Vercel rewrites "/" to "/api"
	if path == "/api" || path == "/api/" {
		path = "/"
	}

	// Vercel rewrites "/shorten" to "/api/shorten"
	if path == "/api/shorten" {
		path = "/shorten"
	}

	// Vercel rewrites "/abc123" to "/api/abc123"
	if strings.HasPrefix(path, "/api/") {
		path = strings.TrimPrefix(path, "/api")
	}

	switch {
	case path == "/" && r.Method == http.MethodGet:
		showHome(w)

	case path == "/shorten" && r.Method == http.MethodPost:
		shortenURL(w, r)

	case r.Method == http.MethodGet:
		redirectURL(w, r, path)

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func showHome(w http.ResponseWriter) {

	tmpl, err := template.ParseFS(templateFiles, "views/index.html")

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = tmpl.Execute(w, nil)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func shortenURL(w http.ResponseWriter, r *http.Request) {

	originalURL := r.FormValue("url")

	if originalURL == "" {
		http.Error(w, "URL not provided", http.StatusBadRequest)
		return
	}

	if !strings.HasPrefix(originalURL, "http://") &&
		!strings.HasPrefix(originalURL, "https://") {

		originalURL = "https://" + originalURL
	}

	hash := sha256.Sum256([]byte(originalURL))
	shortURL := hex.EncodeToString(hash[:])[:8]

	store.Lock()
	store.data[shortURL] = originalURL
	store.Unlock()

	data := map[string]string{
		"ShortURL": shortURL,
	}

	tmpl, err := template.ParseFS(templateFiles, "views/shorten.html")

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = tmpl.Execute(w, data)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func redirectURL(w http.ResponseWriter, r *http.Request, path string) {

	shortURL := strings.TrimPrefix(path, "/")

	if shortURL == "" {
		http.NotFound(w, r)
		return
	}

	store.RLock()
	originalURL, exists := store.data[shortURL]
	store.RUnlock()

	if !exists {
		http.Error(
			w,
			"Short URL not found. It may have expired.",
			http.StatusNotFound,
		)
		return
	}

	http.Redirect(w, r, originalURL, http.StatusTemporaryRedirect)
}