package controllers

import (
	"database/sql"
	"html/template"
	"net/http"
	"strings"

	"github.com/Emmanuel-Wantua/url-shortner-app.git/internal/db"
	"github.com/Emmanuel-Wantua/url-shortner-app.git/internal/url"
)

//func GenerateShortID() string {
//	randomInt := rand.Intn(10)
//	randomInt2 := rand.Intn(10)
//	randomInt3 := rand.Intn(10)
//
//	randomLetter := 'a' + byte(rand.Intn(26))
//	randomLetter2 := 'a' + byte(rand.Intn(26))
//	randomLetter3 := 'a' + byte(rand.Intn(26))
//
//	return fmt.Sprintf("%c%c%c%d%d%d", randomLetter, randomLetter2, randomLetter3, randomInt, randomInt2, randomInt3)
//}

func Shorten(lite *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		originalUrl := r.FormValue("url")
		if originalUrl == "" {
			http.Error(w, "Url not provided", http.StatusBadRequest)
			return
		}

		if !strings.HasPrefix(originalUrl, "http://") && !strings.HasPrefix(originalUrl, "https://") {
			originalUrl = "https://" + originalUrl
		}

		//generatedID := GenerateShortID()

		//shortUrl := map[string]string{
		//	generatedID: originalUrl,
		//}

		shortURL := url.Shorten(originalUrl)

		if err := db.StoreURL(lite, shortURL, originalUrl); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		data := map[string]string{
			"ShortURL": shortURL,
		}

		t, err := template.ParseFiles("internal/views/shorten.html")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if err = t.Execute(w, data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func Proxy(lite *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		shortURL := r.URL.Path[1:]
		if shortURL == "" {
			http.Error(w, "Url not provided", http.StatusBadRequest)
			return
		}
		origUrl, err := db.GetOriginalURL(lite, shortURL)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		http.Redirect(w, r, origUrl, http.StatusPermanentRedirect)
	}
}
