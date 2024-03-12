package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"text/template"
	"twilu/cmd/api/handler"
	"twilu/internal/cfg"
	"twilu/internal/controller"
	"twilu/internal/database"
)

func main() {
	store := cfg.InitializeSessionStore()
	db, err := database.New()
	if err != nil {
		log.Fatal(err)
	}
	userController := controller.NewUserController(db)
	itemController := controller.NewItemController(db)
	folderController := controller.NewFolderController(db)

	userHandler := handler.NewUserHandler(store, userController)
	itemHandler := handler.NewItemHandler(store, itemController)
	folderHandler := handler.NewFolderHandler(store, folderController)

	mux := http.NewServeMux()
	mux.Handle("/internal/web", http.StripPrefix("/internal/web", http.FileServer(http.Dir("./internal/web"))))

	// html pages
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		sess, _ := store.Get(r, "twilu-cookie")
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
		sess, _ := store.Get(r, "twilu-cookie")
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
		sess, _ := store.Get(r, "twilu-cookie")
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
		sess, _ := store.Get(r, "twilu-cookie")
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
		sess, _ := store.Get(r, "twilu-cookie")
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
	mux.HandleFunc("POST /api/signup", userHandler.SignUp)
	mux.HandleFunc("POST /api/login", userHandler.Login)
	mux.HandleFunc("POST /api/logout", userHandler.Logout)
	mux.HandleFunc("GET /api/user/folders", userHandler.GetFolders)
	mux.HandleFunc("POST /api/folder/create", folderHandler.CreateFolder)
	mux.HandleFunc("GET /api/folder/{id}", folderHandler.GetFolder)
	mux.HandleFunc("DELETE /api/folder/{id}", folderHandler.DeleteFolder)
	mux.HandleFunc("POST /api/folder/{id}/add", itemHandler.AddItem)
	mux.HandleFunc("DELETE /api/folder/{id}/item/{itemID}", itemHandler.DeleteItem)
	mux.HandleFunc("DELETE /api/user", userHandler.DeleteAccount)
	mux.HandleFunc("GET /api/feed", folderHandler.GetFeed)
	mux.HandleFunc("GET /api/user", userHandler.GetUser)
	mux.HandleFunc("POST /api/password/update", userHandler.UpdatePassword)
	port := os.Getenv("PORT")
	portStr := fmt.Sprintf("0.0.0.0:%s", port)
	log.Fatal(http.ListenAndServe(portStr, mux))
}
