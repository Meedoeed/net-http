package main

import (
	"html/template"
	"log"
	"net/http"
)

var tmpl *template.Template = template.Must(template.ParseFiles(
	"templates/layout.html",
	"templates/login.html",
	"templates/profilelayout.html",
	"templates/profile.html",
))

func ProfileHandler(w http.ResponseWriter, r *http.Request) {
	type PageData struct {
		Title    string
		Error    string
		Username string
	}
	if r.Method == "GET" {
		data := PageData{
			Title:    "Личный кабинет",
			Username: "testuser",
		}
		tmpl.ExecuteTemplate(w, "profilelayout.html", data)
	} else {
		http.NotFound(w, r)
		return
	}
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	type PageData struct {
		Title string
		Error string
	}
	if r.Method == "GET" {
		data := PageData{
			Title: "Вход",
			Error: "",
		}
		err := tmpl.ExecuteTemplate(w, "layout.html", data)
		if err != nil {
			http.Error(w, "Template error", 500)
			return
		}
	}
	if r.Method == "POST" {
		w.Header().Add("Content-type", "text/html")
		err := r.ParseForm()
		if err != nil {
			http.Error(w, "Form is not valid 1", http.StatusBadRequest)
			return
		}
		login := r.FormValue("login")
		password := r.FormValue("password")
		if password == "" || login == "" {
			http.Error(w, "Form is not valid 2", http.StatusBadRequest)
			return
		}
		response := "Logged as " + login
		w.Write([]byte(response))
	}
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("[INFO] METHOD PATH %s %s (%s)", r.Method, r.URL.Path, r.RemoteAddr)
		next.ServeHTTP(w, r)
	})
}

func recoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				log.Printf("Panic recovered: %s", err)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/login", LoginHandler)
	mux.HandleFunc("/profile", ProfileHandler)
	MuxModified := recoveryMiddleware(loggingMiddleware(mux))
	http.ListenAndServe(":8080", MuxModified)
}
