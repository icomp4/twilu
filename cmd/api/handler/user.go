package handler

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/sessions"
	"html/template"
	"io"
	"net/http"
	"path/filepath"
	"twilu/internal/controller"
	"twilu/internal/model"
	"twilu/internal/util"
)

type UserHandler struct {
	store      *sessions.CookieStore
	controller *controller.UserController
}

func NewUserHandler(store *sessions.CookieStore, controller *controller.UserController) *UserHandler {
	return &UserHandler{
		store:      store,
		controller: controller}
}

func (uh *UserHandler) SignUp(w http.ResponseWriter, r *http.Request) {
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
	user.ProfilePicture = "https://www.testhouse.net/wp-content/uploads/2021/11/default-avatar.jpg"
	if err := uh.controller.CreateAccount(user); err != nil {
		io.WriteString(w, "Email or username already in use")
		return
	}
	w.Header().Set("HX-Redirect", "/login")
	w.WriteHeader(http.StatusAccepted)
}
func (uh *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	sess, err := uh.store.Get(r, "twilu-cookie")
	if err != nil {
		http.Error(w, "Failed to retrieve session", http.StatusInternalServerError)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Error parsing the form", http.StatusInternalServerError)
		return
	}

	var user model.User
	user.Username = r.PostFormValue("username")
	user.Password = r.PostFormValue("password")

	userInfo, err := uh.controller.SignIn(user)
	if err != nil {
		io.WriteString(w, "Incorrect login info")
		return
	}
	sess.Values["userID"] = int(userInfo.ID)
	sess.Values["authenticated"] = true
	sess.Save(r, w)
	w.Header().Set("HX-Redirect", "/main")
	w.WriteHeader(http.StatusAccepted)
}
func (uh *UserHandler) Logout(w http.ResponseWriter, r *http.Request) {
	sess, err := uh.store.Get(r, "twilu-cookie")
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
func (uh *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	sess, err := uh.store.Get(r, "twilu-cookie")
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

	user, err := uh.controller.GetUserByID(userIDInt)
	if err != nil {
		http.Error(w, "Unable to get folders", http.StatusInternalServerError)
		return
	}

	tmplPath := filepath.Join("./internal/web/templates", "account.html")
	tmpl, err := template.ParseFiles(tmplPath)
	if err != nil {
		http.Error(w, "Unable to load template", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	err = tmpl.Execute(w, user)
	if err != nil {
		http.Error(w, "Unable to execute template", http.StatusInternalServerError)
		return
	}
}
func (uh *UserHandler) GetFolders(w http.ResponseWriter, r *http.Request) {
	sess, err := uh.store.Get(r, "twilu-cookie")
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

	folders, err := uh.controller.GetUserFoldersByID(userIDInt)
	if err != nil {
		http.Error(w, "Unable to get folders", http.StatusInternalServerError)
		return
	}

	tmplPath := filepath.Join("./internal/web/templates", "folders.html")
	tmpl, err := template.ParseFiles(tmplPath)
	if err != nil {
		http.Error(w, "Unable to load template", http.StatusInternalServerError)
		return
	}

	foldersJSON, err := json.Marshal(folders)
	if err != nil {
		http.Error(w, "Unable to marshal folders", http.StatusInternalServerError)
		return
	}

	data := struct {
		Folders     []model.Folder
		FoldersJSON template.JS
		HasFolders  bool
	}{
		Folders:     folders,
		FoldersJSON: template.JS(foldersJSON),
		HasFolders:  len(folders) > 0,
	}
	w.Header().Set("Content-Type", "text/html")
	err = tmpl.Execute(w, data)
	if err != nil {
		http.Error(w, "Unable to execute template", http.StatusInternalServerError)
		return
	}
}
func (uh *UserHandler) UpdatePassword(w http.ResponseWriter, r *http.Request) {
	sess, err := uh.store.Get(r, "twilu-cookie")
	if err != nil {
		http.Error(w, "Bad session", http.StatusBadGateway)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Error parsing the form", http.StatusInternalServerError)
		return
	}
	currentPw := r.PostFormValue("currentPassword")
	newPw := r.PostFormValue("newPassword")
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
	if currentPw == "" || newPw == "" {
		fmt.Fprint(w, "<div class='error'>Fields must not be blank</div>")
		return
	}

	err2 := uh.controller.UpdatePassword(userIDInt, currentPw, newPw)
	if err2 != nil {
		fmt.Fprint(w, "<div class='error'>Unable to update password.</div>")
		return
	}
	w.Header().Set("HX-Redirect", "/account")
	w.WriteHeader(http.StatusAccepted)
	fmt.Fprint(w, "<div class='success'>Password successfully updated.</div>")

}
func (uh *UserHandler) DeleteAccount(w http.ResponseWriter, r *http.Request) {
	sess, err := uh.store.Get(r, "twilu-cookie")
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

	err2 := uh.controller.DeleteAccount(userIDInt)
	if err2 != nil {
		http.Error(w, "failed to delete account", http.StatusBadRequest)
		return
	}
	sess.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   true,
	}
	sess.Values["authenticated"] = false
	err3 := sess.Save(r, w)
	if err3 != nil {
		http.Error(w, "Failed to save sess", http.StatusInternalServerError)
		return
	}
}
