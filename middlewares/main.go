package main

import (
	"log"
	"net/http"
	"time"
)

type timeHandler struct{}

func homeHandler(w http.ResponseWriter, req *http.Request) {
	if req.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
	if req.URL.Path != "/" {
		http.NotFound(w, req)
		return
	}
	w.Header().Set("Content-type", "text/html; charset=8-utf")
	w.Write([]byte("<h1>Home Page</h1><p>Welcome to the structured server!</p>"))
}

func (t timeHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if req.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
	if req.URL.Path != "/api/time" {
		http.NotFound(w, req)
		return
	}
	w.Header().Set("Content-type", "text/plain; charset=8-utf")
	w.Write([]byte(time.Now().UTC().Format(time.RFC3339)))
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		log.Printf("[INFO] METHOD PATH %s %s", req.Method, req.URL.Path)
		next.ServeHTTP(w, req)
	})
}

func recoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if err := recover(); err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			log.Printf("Panic recovered: %s", err)
		}
		next.ServeHTTP(w, req)
	})
}

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("/", homeHandler)
	mux.Handle("/api/time", timeHandler{})

	SuperMux := recoveryMiddleware(loggingMiddleware(mux))

	err := http.ListenAndServe(":8080", SuperMux)
	if err != nil {
		log.Fatal(err)
	}
}
