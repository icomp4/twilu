package handler

import (
	"github.com/gorilla/sessions"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"strconv"
	"twilu/internal/controller"
	"twilu/internal/model"
)

type FolderHandler struct {
	store      *sessions.CookieStore
	controller *controller.FolderController
}

func NewFolderHandler(store *sessions.CookieStore, controller *controller.FolderController) *FolderHandler {
	return &FolderHandler{
		store:      store,
		controller: controller}
}

func (h *FolderHandler) GetFeed(w http.ResponseWriter, r *http.Request) {
	sess, err := h.store.Get(r, "twilu-cookie")
	if err != nil {
		http.Error(w, "Bad session", http.StatusBadGateway)
		return
	}

	_, ok := sess.Values["userID"]
	if !ok {
		http.Error(w, "User ID not found in session", http.StatusBadRequest)
		return
	}

	folders, err := h.controller.GetFeed()
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
func (h *FolderHandler) GetFolder(w http.ResponseWriter, r *http.Request) {
	_, err := h.store.Get(r, "twilu-cookie")
	if err != nil {
		http.Error(w, "Bad session", http.StatusBadGateway)
		return
	}
	folderID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "unable to find folder", http.StatusBadRequest)
		return
	}
	folder, err := h.controller.GetFolder(folderID)
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
func (h *FolderHandler) CreateFolder(w http.ResponseWriter, r *http.Request) {
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

	sess, err := h.store.Get(r, "twilu-cookie")
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

	create := h.controller.CreateFolder(folder, userIDInt)
	if create != nil {
		http.Error(w, "Failed to create folder", http.StatusBadRequest)
		return
	}
	w.Header().Set("HX-Redirect", "/main")
	w.WriteHeader(http.StatusAccepted)
}
func (h *FolderHandler) DeleteFolder(w http.ResponseWriter, r *http.Request) {
	sess, err := h.store.Get(r, "twilu-cookie")
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
	if err := h.controller.DeleteFolder(folderID, userIDInt); err != nil {
		http.Error(w, "failed to delete folder", http.StatusBadGateway)
		return
	}
	w.Header().Set("HX-Redirect", "/main")
	w.WriteHeader(http.StatusAccepted)
}
