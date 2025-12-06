package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type timeHandler struct{}

type userHandler struct {
}

type loginHandler struct {
}

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (l loginHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if req.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	req.Body = http.MaxBytesReader(w, req.Body, 1<<10)
	if req.URL.Path == "/login-form" {
		w.Header().Add("Content-type", "text/html")
		err := req.ParseForm()
		if err != nil {
			http.Error(w, "Internal error", http.StatusInternalServerError)
		}
		login := req.FormValue("username")
		password := req.FormValue("password")
		if login == "" || password == "" {
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}
		response := "<h1>Hello" + login + "</h1><p>Your password is hidden.</p>"
		w.Write([]byte(response))
	} else if req.URL.Path == "/api/login" {
		w.Header().Add("Content-type", "application/json")
		var r loginRequest
		if err := json.NewDecoder(req.Body).Decode(&r); err != nil {
			http.Error(w, "Invalid Json", http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "success", "user": r.Username})
	} else {
		http.NotFound(w, req)
	}
}

func (u userHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if req.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
	fields := strings.Split(req.URL.Path, "/")
	id := fields[len(fields)-1]
	if _, err := strconv.Atoi(id); err == nil {
		w.Header().Add("Content-type", "text/plain; charset=utf-8")
		result := "User ID:" + id
		w.Write([]byte(result))
		return
	} else if id != "" {
		http.Error(w, "400 - Bad request", http.StatusBadRequest)
		return
	} else {
		http.NotFound(w, req)
	}
}

func homeHandler(w http.ResponseWriter, req *http.Request) {
	if req.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if req.URL.Path != "/" {
		http.NotFound(w, req)
		return
	}
	w.Header().Set("Content-type", "text/html; charset=utf-8")
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
	w.Header().Set("Content-type", "text/plain; charset=utf-8")
	w.Write([]byte(time.Now().UTC().Format(time.RFC3339)))
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		log.Printf("[INFO] METHOD PATH %s %s (%s)", req.Method, req.URL.Path, req.RemoteAddr)
		next.ServeHTTP(w, req)
	})
}

func recoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				log.Printf("Panic recovered: %s", err)
			}
		}()
		next.ServeHTTP(w, req)
	})
}

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("/", homeHandler)
	mux.Handle("/api/time", timeHandler{})
	mux.Handle("/users/", userHandler{})
	mux.Handle("/login-form", loginHandler{})
	mux.Handle("/api/login", loginHandler{})

	SuperMux := recoveryMiddleware(loggingMiddleware(mux))

	err := http.ListenAndServe(":8080", SuperMux)
	if err != nil {
		log.Fatal(err)
	}
}
