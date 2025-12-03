package main

import (
	"log"
	"net/http"
)

func greeting(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Добро пожаловать на главную страницу"))
}

func hello(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	if name == "" {
		name += "незнакомец"
	}
	response := "Привет," + name + "!"
	w.Write([]byte(response))
}

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("/", greeting)
	mux.HandleFunc("/hello", hello)

	err := http.ListenAndServe(":8080", mux)
	if err != nil {
		log.Fatal("server was stopped")
	}
}
