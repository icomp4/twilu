package main

import (
	"log"
	"net/http"
	"text/template"
	"twilu/database"
	"twilu/handler"
)

func main() {
	database.StartDB()
	handler.Init()
	mux := http.NewServeMux()
	mux.Handle("/client/", http.StripPrefix("/client/", http.FileServer(http.Dir("./client"))))

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		sess, _ := handler.Store.Get(r, "twilu-cookie")
		if auth, ok := sess.Values["authenticated"].(bool); ok && auth {
			http.Redirect(w, r, "/main", http.StatusFound)
			return
		}
		templates := template.Must(template.ParseFiles("client/login.html"))
		if err := templates.ExecuteTemplate(w, "login.html", nil); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})
	mux.HandleFunc("/signup", func(w http.ResponseWriter, r *http.Request) {
		templates := template.Must(template.ParseFiles("client/signup.html"))
		if err := templates.ExecuteTemplate(w, "signup.html", nil); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})
	mux.HandleFunc("/main", func(w http.ResponseWriter, r *http.Request) {
		sess, _ := handler.Store.Get(r, "twilu-cookie")
		if auth, ok := sess.Values["authenticated"].(bool); !ok || !auth {
			http.Redirect(w, r, "/", http.StatusFound)
			return
		}
		templates := template.Must(template.ParseFiles("client/main.html"))
		if err := templates.ExecuteTemplate(w, "main.html", nil); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})
	mux.HandleFunc("/folder/{id}", func(w http.ResponseWriter, r *http.Request) {
		sess, _ := handler.Store.Get(r, "twilu-cookie")
		if auth, ok := sess.Values["authenticated"].(bool); !ok || !auth {
			http.Redirect(w, r, "/", http.StatusFound)
			return
		}

		templates := template.Must(template.ParseFiles("client/folderPage.html"))
		if err := templates.ExecuteTemplate(w, "folderPage.html", nil); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})
	mux.HandleFunc("POST /api/signup", handler.SignUp)
	mux.HandleFunc("POST /api/login", handler.LogIn)
	mux.HandleFunc("POST /api/logout", handler.Logout)
	mux.HandleFunc("GET /api/user/folders", handler.GetFolders)
	mux.HandleFunc("POST /api/folder/create", handler.CreateFolder)
	mux.HandleFunc("GET /api/folder/{id}", handler.GetFolder)

	log.Fatal(http.ListenAndServe(":8080", mux))
}
