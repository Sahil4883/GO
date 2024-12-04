package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strings"
	"sync"
	"time"
)

type URLShortener struct {
	mu    sync.RWMutex
	store map[string]string
}

var urlShortener = URLShortener{
	store: make(map[string]string),
}

func generateShortURL() string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	const length = 6
	rand.Seed(time.Now().UnixNano())

	shortURL := strings.Builder{}
	for i := 0; i < length; i++ {
		shortURL.WriteByte(charset[rand.Intn(len(charset))])
	}

	return shortURL.String()
}

// ShortenURL handles the creation of a short URL.
func ShortenURL(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Only POST requests are allowed", http.StatusMethodNotAllowed)
		return
	}

	type RequestBody struct {
		URL string `json:"url"`
	}

	var requestBody RequestBody
	err := json.NewDecoder(r.Body).Decode(&requestBody)
	if err != nil || requestBody.URL == "" {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	shortURL := generateShortURL()

	urlShortener.mu.Lock()
	urlShortener.store[shortURL] = requestBody.URL
	urlShortener.mu.Unlock()

	response := map[string]string{
		"short_url": fmt.Sprintf("http://localhost:8080/%s", shortURL),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Redirect handles the redirection from short URL to the original URL.
func Redirect(w http.ResponseWriter, r *http.Request) {
	shortURL := strings.TrimPrefix(r.URL.Path, "/")

	urlShortener.mu.RLock()
	originalURL, exists := urlShortener.store[shortURL]
	urlShortener.mu.RUnlock()

	if !exists {
		http.Error(w, "URL not found", http.StatusNotFound)
		return
	}

	http.Redirect(w, r, originalURL, http.StatusFound)
}

func main() {
	http.HandleFunc("/shorten", ShortenURL)
	http.HandleFunc("/", Redirect)

	fmt.Println("Server is running on port 8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
