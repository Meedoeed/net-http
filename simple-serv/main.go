package main

import (
	"log"
	"net/http"
)

func welcome(w http.ResponseWriter, req *http.Request) {
	if req.Method != "GET" {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
	if req.URL.Path != "/" {
		http.NotFound(w, req)
		return
	}
	w.Header().Add("Content-type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("<h1>Welcome</h1> \n <h2>This is the home page</h2>"))
}

func NotAllowedHandler(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusMethodNotAllowed)
	w.Write([]byte("<h1>405 - Method Not Allowed</h1>"))
}

func hello(w http.ResponseWriter, req *http.Request) {
	if req.URL.Path != "/hello" {
		http.NotFound(w, req)
		return
	}
	if req.Method != "GET" {
		NotAllowedHandler(w, req)
		return
	}
	w.Header().Add("Content-type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Hello, World!"))
}

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("/", welcome)
	mux.HandleFunc("/hello", hello)
	err := http.ListenAndServe(":8080", mux)
	if err != nil {
		log.Fatal("err")
	}
}
