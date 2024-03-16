package controller

import (
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"log"
	"strings"
	"twilu/internal/model"
	"twilu/internal/util"
)

// UserController handles operations on folders.
type UserController struct {
	DB *gorm.DB
}

// NewUserController creates a new instance of UserController.
func NewUserController(db *gorm.DB) *UserController {
	return &UserController{DB: db}
}
func (uc *UserController) CreateAccount(user model.User) error {
	user.Username = strings.ToLower(user.Username)
	password, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Println(err)
		return err
	}
	user.Password = string(password)
	if err := uc.DB.Create(&user).Error; err != nil {
		return err
	}
	return nil
}
func (uc *UserController) DeleteAccount(id int) error {
	if err := uc.DB.Unscoped().Exec("DELETE FROM user_folders WHERE folder_id IN (SELECT id FROM folders WHERE owner = ?)", id).Error; err != nil {
		return err
	}

	if err := uc.DB.Unscoped().Exec("DELETE FROM folder_contributors WHERE folder_id IN (SELECT id FROM folders WHERE owner = ?)", id).Error; err != nil {
		return err
	}
	if err := uc.DB.Unscoped().Where("owner_id = ?", id).Delete(&model.Item{}).Error; err != nil {
		return err
	}

	if err := uc.DB.Unscoped().Where("owner = ?", id).Delete(&model.Folder{}).Error; err != nil {
		return err
	}

	if err := uc.DB.Unscoped().Where("id = ?", id).Delete(&model.User{}).Error; err != nil {
		return err
	}

	return nil
}

func (uc *UserController) SignIn(user model.User) (model.User, error) {
	var userLookUp model.User
	err := uc.DB.Preload("Folders").Find(&userLookUp, "username = ?", user.Username).Error
	if err != nil {
		return model.User{}, err
	}
	err2 := bcrypt.CompareHashAndPassword([]byte(userLookUp.Password), []byte(user.Password))
	if err2 != nil {
		return model.User{}, err2
	}
	return userLookUp, nil
}
func (uc *UserController) GetUserByID(userID int) (model.User, error) {
	var user model.User
	if err := uc.DB.First(&user, userID).Error; err != nil {
		return model.User{}, err
	}
	return user, nil
}
func (uc *UserController) GetUserFoldersByID(userID int) ([]model.Folder, error) {
	var folders []model.Folder
	if err := uc.DB.Model(&model.Folder{}).Where("owner = ?", userID).Order("created_at DESC").Find(&folders).Error; err != nil {
		return []model.Folder{}, err
	}
	return folders, nil
}
func (uc *UserController) UpdatePassword(userID int, currentPw string, newPw string) error {
	var user model.User
	if err := uc.DB.First(&user, userID).Error; err != nil {
		return err
	}
	if util.PasswordIsValid(newPw) == false {
		return fmt.Errorf("invalid password")
	}
	if newPw == currentPw {
		return fmt.Errorf("new password can't be the same as the old one")
	}
	err2 := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(currentPw))
	if err2 != nil {
		return fmt.Errorf("incorrect password")
	}
	password, err := bcrypt.GenerateFromPassword([]byte(newPw), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	user.Password = string(password)
	if err := uc.DB.Save(&user).Error; err != nil {
		return err
	}
	return nil
}
