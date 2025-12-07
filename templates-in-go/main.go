package main

import (
	"html/template"
	"net/http"
)

var tmpl *template.Template = template.Must(template.ParseFiles(
	"templates/layout.html",
	"templates/login.html",
))

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	type PageData struct {
		Title string
		Error string
	}
	if r.Method == "GET" {
		data := PageData{
			Title: "Вход в систему",
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

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/login", LoginHandler)
	http.ListenAndServe(":8080", mux)
}
