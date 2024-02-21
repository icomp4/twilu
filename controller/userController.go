package controller

import (
	"log"
	"strings"
	"twilu/database"
	"twilu/model"

	"golang.org/x/crypto/bcrypt"
)

func CreateAccount(user model.User) error {
	user.Username = strings.ToLower(user.Username)
	password, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Println(err)
		return err
	}
	user.Password = string(password)
	if err := database.DB.Create(&user).Error; err != nil {
		return err
	}
	return nil
}

func DeleteAccount(id int) error {
	var user model.User
	if err := database.DB.Unscoped().Delete(&user, id).Error; err != nil { // unscoped actually deletes the record, instead of soft deleting
		return err
	}
	return nil
}

func SignIn(user model.User) (model.User, error) {
	var userLookUp model.User
	err := database.DB.Preload("Folders").Find(&userLookUp, "username = ?", user.Username).Error
	if err != nil {
		return model.User{}, err
	}
	err2 := bcrypt.CompareHashAndPassword([]byte(userLookUp.Password), []byte(user.Password))
	if err2 != nil {
		return model.User{}, err2
	}
	return userLookUp, nil
}

func GetUserByID(userID int) (model.User, error) {
	var user model.User
	if err := database.DB.First(&user, userID).Error; err != nil {
		return model.User{}, err
	}
	return user, nil
}
func GetUserFoldersByID(userID int) ([]model.Folder, error) {
	var folders []model.Folder
	if err := database.DB.Model(&model.Folder{}).Where("owner = ?", userID).Find(&folders).Error; err != nil {
		return []model.Folder{}, err
	}
	return folders, nil
}
