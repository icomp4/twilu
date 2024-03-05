package handler

import (
	"fmt"
	"github.com/gorilla/sessions"
	"net/http"
	"strconv"
	"twilu/internal/controller"
	"twilu/internal/model"
)

type ItemHandler struct {
	store      *sessions.CookieStore
	controller *controller.ItemController
}

func NewItemHandler(store *sessions.CookieStore, controller *controller.ItemController) *ItemHandler {
	return &ItemHandler{
		store:      store,
		controller: controller}
}

func (ih *ItemHandler) AddItem(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Error parsing the form", http.StatusInternalServerError)
		return
	}

	var item model.Item
	item.Name = r.PostFormValue("itemName")
	item.URL = r.PostFormValue("itemUrl")

	sess, err := ih.store.Get(r, "twilu-cookie")
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
	if err := ih.controller.AddItemToFolder(folderID, item, userIDInt); err != nil {
		http.Error(w, "unable to convert id", http.StatusBadGateway)
		return
	}
	url := "/folder/" + fmt.Sprint(folderID)
	w.Header().Set("HX-Redirect", url)
	w.WriteHeader(http.StatusAccepted)
}
func (ih *ItemHandler) DeleteItem(w http.ResponseWriter, r *http.Request) {
	sess, err := ih.store.Get(r, "twilu-cookie")
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
	if err := ih.controller.DeleteItem(folderID, userIDInt, itemID); err != nil {
		http.Error(w, "unable to delete item", http.StatusBadGateway)
		return
	}
	url := "/folder/" + fmt.Sprint(folderID)
	w.Header().Set("HX-Redirect", url)
	w.WriteHeader(http.StatusAccepted)
}
