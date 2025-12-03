package main

import "net/http"

func greeting(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Добро пожаловать на главную страницу"))
}

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("/", greeting)

	http.ListenAndServe(":8080", mux)
}
