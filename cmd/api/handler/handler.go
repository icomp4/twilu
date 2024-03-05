package handler

import (
	"fmt"
	"github.com/gorilla/sessions"
	"github.com/joho/godotenv"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"twilu/internal/controller"
	"twilu/internal/model"
	"twilu/internal/util"
)

var Store *sessions.CookieStore

func Init() {
	if err := godotenv.Load(".env"); err != nil {
		log.Print("No .env file found")
	}
	sessionKey := os.Getenv("SESSION_KEY")
	if sessionKey == "" {
		log.Fatal("SESSION_KEY is not set")
	}
	Store = sessions.NewCookieStore([]byte(sessionKey))
	Store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   3600 * 24,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	}
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
	user.ProfilePicture = "https://www.testhouse.net/wp-content/uploads/2021/11/default-avatar.jpg"
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

	tmplPath := filepath.Join("./internal/web/templates", "folders.html")
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
func GetUser(w http.ResponseWriter, r *http.Request) {
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

	user, err := controller.GetUserByID(userIDInt)
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
func UpdatePassword(w http.ResponseWriter, r *http.Request) {
	sess, err := Store.Get(r, "twilu-cookie")
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

	err2 := controller.UpdatePassword(userIDInt, currentPw, newPw)
	if err2 != nil {
		fmt.Fprint(w, "<div class='error'>Unable to update password.</div>")
		return
	}
	fmt.Fprint(w, "<div class='success'>Password successfully updated.</div>")

	w.Header().Set("HX-Redirect", "/account")
	w.WriteHeader(http.StatusAccepted)

}
func GetFeed(w http.ResponseWriter, r *http.Request) {
	sess, err := Store.Get(r, "twilu-cookie")
	if err != nil {
		http.Error(w, "Bad session", http.StatusBadGateway)
		return
	}

	_, ok := sess.Values["userID"]
	if !ok {
		http.Error(w, "User ID not found in session", http.StatusBadRequest)
		return
	}

	folders, err := controller.GetFeed()
	if err != nil {
		http.Error(w, "Unable to get folders", http.StatusInternalServerError)
		return
	}

	tmplPath := filepath.Join("./internal/web/templates", "social.html")
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
func GetFolder(w http.ResponseWriter, r *http.Request) {
	_, err := Store.Get(r, "twilu-cookie")
	if err != nil {
		http.Error(w, "Bad session", http.StatusBadGateway)
		return
	}
	folderID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "unable to find folder", http.StatusBadRequest)
		return
	}
	folder, err := controller.GetFolder(folderID)
	if err != nil {
		http.Error(w, "unable to find folder", http.StatusBadRequest)
		return
	}
	type TemplateData struct {
		Folder model.Folder // Assuming Folder is the struct type
	}
	tmplData := TemplateData{Folder: folder}
	tmplPath := filepath.Join("./internal/web/templates", "folder.html")
	tmpl, err := template.ParseFiles(tmplPath)
	if err != nil {
		http.Error(w, "Unable to load template", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html")
	err = tmpl.Execute(w, tmplData)
	if err != nil {
		log.Println("Unable to execute template")
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
func DeleteFolder(w http.ResponseWriter, r *http.Request) {
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

	folderID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "unable to convert id", http.StatusBadGateway)
		return
	}
	if err := controller.DeleteFolder(folderID, userIDInt); err != nil {
		http.Error(w, "failed to delete folder", http.StatusBadGateway)
		return
	}
	w.Header().Set("HX-Redirect", "/main")
	w.WriteHeader(http.StatusAccepted)
}
func AddItem(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Error parsing the form", http.StatusInternalServerError)
		return
	}

	var item model.Item
	item.Name = r.PostFormValue("itemName")
	item.URL = r.PostFormValue("itemUrl")

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
	folderID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "unable to convert id", http.StatusBadGateway)
		return
	}
	if err := controller.AddItemToFolder(folderID, item, userIDInt); err != nil {
		http.Error(w, "unable to convert id", http.StatusBadGateway)
		return
	}
	url := "/folder/" + fmt.Sprint(folderID)
	w.Header().Set("HX-Redirect", url)
	w.WriteHeader(http.StatusAccepted)
}
func DeleteItem(w http.ResponseWriter, r *http.Request) {
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

	folderID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "unable to convert id", http.StatusBadGateway)
		return
	}
	itemID, err := strconv.Atoi(r.PathValue("itemID"))
	if err != nil {
		http.Error(w, "unable to convert id", http.StatusBadGateway)
		return
	}
	if err := controller.DeleteItem(folderID, userIDInt, itemID); err != nil {
		http.Error(w, "unable to delete item", http.StatusBadGateway)
		return
	}
	url := "/folder/" + fmt.Sprint(folderID)
	w.Header().Set("HX-Redirect", url)
	w.WriteHeader(http.StatusAccepted)
}
