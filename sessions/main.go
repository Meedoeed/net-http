package main

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"
)

var tmpl *template.Template = template.Must(template.ParseFiles(
	"templates/layout.html",
	"templates/login.html",
	"templates/profilelayout.html",
	"templates/profile.html",
))

type UserSession struct {
	username string
	avatar   string
}

var (
	sessions   = make(map[string]UserSession)
	sessionMux sync.RWMutex
)

func generateSesssion() string {
	b := make([]byte, 48)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}

func setSession(w http.ResponseWriter, username string) {
	id := generateSesssion()
	sessionMux.Lock()
	sessions[id] = UserSession{username: username, avatar: ""}
	sessionMux.Unlock()
	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    id,
		Path:     "/",
		HttpOnly: true,
		MaxAge:   24 * 3600,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
	})
}

func getSessionFull(r *http.Request) (session UserSession, ok bool) {
	cookie, err := r.Cookie("session_id")
	session = UserSession{username: "", avatar: ""}
	if err != nil {
		return session, false
	}
	sessionMux.RLock()
	session, ok = sessions[cookie.Value]
	sessionMux.RUnlock()
	if !ok {
		return session, false
	}
	return session, ok
}

func getSession(r *http.Request) (useename string, ok bool) {
	cookie, err := r.Cookie("session_id")
	username := ""
	if err != nil {
		return username, false
	}
	sessionMux.RLock()
	session, ok := sessions[cookie.Value]
	sessionMux.RUnlock()
	if ok {
		username = session.username
	}
	return username, ok
}

func uploaderHandler(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 5<<20)
	file, header, err := r.FormFile("avatar")
	if err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}
	defer file.Close()
	ext := filepath.Ext(header.Filename)
	safename := fmt.Sprintf("%d%s", time.Now().UnixNano(), ext)
	savepath := filepath.Join("uploads", safename)
	out, err := os.Create(savepath)

	if err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}
	session, ok := getSessionFull(r)
	if !ok {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}
	session.avatar = "/" + filepath.ToSlash(savepath)
	cookie, err := r.Cookie("session_id")
	if err != nil {
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}
	defer out.Close()
	io.Copy(out, file)
	sessionMux.Lock()
	sessions[cookie.Value] = session
	sessionMux.Unlock()
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func clearSession(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session_id")
	if err == nil {
		sessionMux.Lock()
		delete(sessions, cookie.Value)
		sessionMux.Unlock()
	}
	http.SetCookie(w, &http.Cookie{
		Name:   "session_id",
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	})
}

type PageData struct {
	Title    string
	Error    string
	Username string
	Avatar   string
}

func ProfileHandler(w http.ResponseWriter, r *http.Request) {
	if session, ok := getSessionFull(r); ok { // temporary
		if r.Method == "GET" {
			data := PageData{
				Title:    "Личный кабинет",
				Username: session.username,
				Avatar:   session.avatar,
			}
			tmpl.ExecuteTemplate(w, "profilelayout.html", data)
		} else {
			http.NotFound(w, r)
			return
		}
	} else {
		http.Redirect(w, r, "/login", http.StatusFound) // temporary
	}
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
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
}

func validationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/login" || r.Method != http.MethodPost {
			next.ServeHTTP(w, r)
			return
		}
		login := r.FormValue("login")
		password := r.FormValue("password")
		err := ""
		title := ""
		data := PageData{
			Title:    title,
			Error:    err,
			Username: login,
		}
		if login == "admin" && password == "secure" {
			data.Title = "Successfully login"
			data.Username = login
			setSession(w, login)
			http.Redirect(w, r, "/profile", http.StatusFound)
			return
		} else {
			data.Title = "Вход"
			data.Error = "Ошибка: Неверный логин или пароль"
			err := tmpl.ExecuteTemplate(w, "layout.html", data)
			if err != nil {
				http.Error(w, "Internal error", http.StatusInternalServerError)
				return
			}
		}
		next.ServeHTTP(w, r)
	})
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

func SuperHandler(w http.ResponseWriter, r *http.Request) {
	if _, ok := getSession(r); ok {
		http.Redirect(w, r, "/profile", http.StatusFound)
	} else {
		http.Redirect(w, r, "/login", http.StatusFound)
	}
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	clearSession(w, r)
	if r.Method == "POST" {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
	} else {
		http.Error(w, "Not allowed method", http.StatusMethodNotAllowed)
	}
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", SuperHandler)
	mux.HandleFunc("/logout", logoutHandler)
	mux.HandleFunc("/login", LoginHandler)
	mux.HandleFunc("/profile", ProfileHandler)
	mux.HandleFunc("/upload-avatar", uploaderHandler)
	mux.Handle("/uploads/", http.StripPrefix("/uploads/", http.FileServer(http.Dir("uploads"))))
	MuxModified := recoveryMiddleware(loggingMiddleware(validationMiddleware(mux)))
	http.ListenAndServe(":8080", MuxModified)
}
