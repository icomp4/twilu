package handler

import (
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"twilu/controller"
	"twilu/model"
	"twilu/util"

	"github.com/gorilla/sessions"
	"github.com/joho/godotenv"
)

var Store *sessions.CookieStore

func Init() {
	if err := godotenv.Load(); err != nil {
		log.Print("No .env file found")
	}
	sessionKey := os.Getenv("SESSION_KEY")
	if sessionKey == "" {
		log.Fatal("SESSION_KEY is not set")
	}
	Store = sessions.NewCookieStore([]byte(sessionKey))
}

func SignUp(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Error parsing the form", http.StatusInternalServerError)
		return
	}

	var user model.User
	user.Username = r.PostFormValue("username")
	user.Email = r.PostFormValue("email")
	user.Password = r.PostFormValue("password")

	if user.Username == "" || user.Password == "" || user.Email == "" {
		io.WriteString(w, "Fields must not be blank")
		return
	}
	if len(user.Username) < 3 {
		io.WriteString(w, "Username must be at least 3 characters long")
		return
	}
	if !util.PasswordIsValid(user.Password) {
		io.WriteString(w, "Please choose a stronger password")
		return
	}

	if err := controller.CreateAccount(user); err != nil {
		io.WriteString(w, "Email or username already in use")
		return
	}
	w.Header().Set("HX-Redirect", "/login")
	w.WriteHeader(http.StatusAccepted)
}

func LogIn(w http.ResponseWriter, r *http.Request) {
	sess, err := Store.Get(r, "twilu-cookie")
	if err != nil {
		io.WriteString(w, "Bad session")
		w.WriteHeader(http.StatusBadGateway)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Error parsing the form", http.StatusInternalServerError)
		return
	}

	var user model.User
	user.Username = r.PostFormValue("username")
	user.Password = r.PostFormValue("password")

	userInfo, err := controller.SignIn(user)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		io.WriteString(w, "Incorrect login info")
		return
	}
	sess.Values["userID"] = int(userInfo.ID)
	sess.Values["authenticated"] = true
	sess.Save(r, w)
	w.Header().Set("HX-Redirect", "/main")
	w.WriteHeader(http.StatusAccepted)
}
func Logout(w http.ResponseWriter, r *http.Request) {
	sess, err := Store.Get(r, "twilu-cookie")
	if err != nil {
		io.WriteString(w, "Bad session")
		w.WriteHeader(http.StatusBadGateway)
		return
	}
	sess.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   true,
	}

	err2 := sess.Save(r, w)
	if err2 != nil {
		http.Error(w, "Failed to save sess", http.StatusInternalServerError)
		return
	}
	fmt.Fprintf(w, `<script>window.location.href = "/";</script>`)
}
func GetFolders(w http.ResponseWriter, r *http.Request) {
	sess, err := Store.Get(r, "twilu-cookie")
	if err != nil {
		http.Error(w, "Bad session", http.StatusBadGateway)
		return
	}

	userID, ok := sess.Values["userID"]
	if !ok {
		http.Error(w, "User ID not found in session", http.StatusBadRequest)
		return
	}

	userIDInt, ok := userID.(int)
	if !ok {
		http.Error(w, "User ID is of invalid type", http.StatusBadRequest)
		return
	}

	folders, err := controller.GetUserFoldersByID(userIDInt)
	if err != nil {
		http.Error(w, "Unable to get folders", http.StatusInternalServerError)
		return
	}

	tmplPath := filepath.Join("templates", "folders.html")
	tmpl, err := template.ParseFiles(tmplPath)
	if err != nil {
		http.Error(w, "Unable to load template", http.StatusInternalServerError)
		return
	}

	data := struct {
		Folders    []model.Folder
		HasFolders bool
	}{
		Folders:    folders,
		HasFolders: len(folders) > 0,
	}

	w.Header().Set("Content-Type", "text/html")
	err = tmpl.Execute(w, data)
	if err != nil {
		http.Error(w, "Unable to execute template", http.StatusInternalServerError)
		return
	}
}
func CreateFolder(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Error parsing the form", http.StatusInternalServerError)
		return
	}

	var folder model.Folder
	folder.Name = r.PostFormValue("folderTitle")
	privacy := r.PostFormValue("isPrivate")
	if privacy == "private" {
		folder.Private = true
	}
	folder.CoverURL = r.PostFormValue("coverUrl")

	sess, err := Store.Get(r, "twilu-cookie")
	if err != nil {
		http.Error(w, "Bad session", http.StatusBadGateway)
		return
	}

	userID, ok := sess.Values["userID"]
	if !ok {
		http.Error(w, "User ID not found in session", http.StatusBadRequest)
		return
	}

	userIDInt, ok := userID.(int)
	if !ok {
		http.Error(w, "User ID is of invalid type", http.StatusBadRequest)
		return
	}

	create := controller.CreateFolder(folder, userIDInt)
	if create != nil {
		http.Error(w, "Failed to create folder", http.StatusBadRequest)
		return
	}
	w.Header().Set("HX-Redirect", "/main")
	w.WriteHeader(http.StatusAccepted)
}
