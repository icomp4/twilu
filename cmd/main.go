package main

import (
	"log"
	"net/http"
	"text/template"
	"twilu/cmd/api/handler"
	"twilu/internal/database"
)

func main() {
	database.StartDB()
	handler.Init()
	mux := http.NewServeMux()
	mux.Handle("/internal/web", http.StripPrefix("/internal/web", http.FileServer(http.Dir("./internal/web"))))
	// html pages
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		sess, _ := handler.Store.Get(r, "twilu-cookie")
		if auth, ok := sess.Values["authenticated"].(bool); ok && auth {
			http.Redirect(w, r, "/main", http.StatusFound)
			return
		}
		templates := template.Must(template.ParseFiles("internal/web/client/login.html"))
		if err := templates.ExecuteTemplate(w, "login.html", nil); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})
	mux.HandleFunc("/signup", func(w http.ResponseWriter, r *http.Request) {
		templates := template.Must(template.ParseFiles("internal/web/client/signup.html"))
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
		templates := template.Must(template.ParseFiles("internal/web/client/main.html"))
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

		templates := template.Must(template.ParseFiles("internal/web/client/folderPage.html"))
		if err := templates.ExecuteTemplate(w, "folderPage.html", nil); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})
	mux.HandleFunc("/social", func(w http.ResponseWriter, r *http.Request) {
		sess, _ := handler.Store.Get(r, "twilu-cookie")
		if auth, ok := sess.Values["authenticated"].(bool); !ok || !auth {
			http.Redirect(w, r, "/", http.StatusFound)
			return
		}

		templates := template.Must(template.ParseFiles("internal/web/client/socialPage.html"))
		if err := templates.ExecuteTemplate(w, "socialPage.html", nil); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})
	mux.HandleFunc("/account", func(w http.ResponseWriter, r *http.Request) {
		sess, _ := handler.Store.Get(r, "twilu-cookie")
		if auth, ok := sess.Values["authenticated"].(bool); !ok || !auth {
			http.Redirect(w, r, "/", http.StatusFound)
			return
		}

		templates := template.Must(template.ParseFiles("internal/web/client/account.html"))
		if err := templates.ExecuteTemplate(w, "account.html", nil); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	// api routes
	mux.HandleFunc("POST /api/signup", handler.SignUp)
	mux.HandleFunc("POST /api/login", handler.LogIn)
	mux.HandleFunc("POST /api/logout", handler.Logout)
	mux.HandleFunc("GET /api/user/folders", handler.GetFolders)
	mux.HandleFunc("POST /api/folder/create", handler.CreateFolder)
	mux.HandleFunc("GET /api/folder/{id}", handler.GetFolder)
	mux.HandleFunc("DELETE /api/folder/{id}", handler.DeleteFolder)
	mux.HandleFunc("POST /api/folder/{id}/add", handler.AddItem)
	mux.HandleFunc("DELETE /api/folder/{id}/item/{itemID}", handler.DeleteItem)
	mux.HandleFunc("GET /api/feed", handler.GetFeed)
	mux.HandleFunc("GET /api/user", handler.GetUser)
	mux.HandleFunc("POST /api/password/update", handler.UpdatePassword)

	log.Fatal(http.ListenAndServe(":8080", mux))
}
