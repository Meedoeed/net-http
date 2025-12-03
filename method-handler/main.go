package main

import (
	"io"
	"log"
	"net/http"
	"strings"
)

func inspectHandler(w http.ResponseWriter, r *http.Request) {
	response := ""
	response += r.Method + "\n" + "ADRESS" + "\n"
	response += r.URL.String() + "\n"
	response += "HEADERS" + "\n"
	for n, v := range r.Header {
		response += n + ": " + strings.Join(v, ", ") + "\n"
	}
	response += "QUERY" + "\n"
	for n, v := range r.URL.Query() {
		response += n + ": " + strings.Join(v, ", ") + "\n"
	}
	if r.Method == "PUT" || r.Method == "POST" {
		contentType := r.Header.Get("Content-Type")
		if strings.Contains(contentType, "application/x-www-form-urlencoded") {
			if err := r.ParseForm(); err == nil {
				if len(r.PostForm) > 0 {
					response += "FORM DATA\n"
					for key, value := range r.PostForm {
						response += key + ": " + strings.Join(value, ", ") + "\n"
					}
				}
			}
		}
		if strings.Contains(contentType, "application/json") {
			defer r.Body.Close()
			body, err := io.ReadAll(r.Body)
			if err != nil {
				log.Fatal("cant read body")
			}
			response += string(body)
		}

		w.Write([]byte(response))
	}
}

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("/inspect", inspectHandler)

	err := http.ListenAndServe(":8080", mux)
	if err != nil {
		log.Fatal("server was stopped")
	}
}
