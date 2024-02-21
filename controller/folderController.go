package controller

import (
	"fmt"
	"twilu/database"
	"twilu/model"

	"gorm.io/gorm"
)

func CreateFolder(folder model.Folder, userID int) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		var user model.User
		if err := tx.First(&user, userID).Error; err != nil {
			return fmt.Errorf("user not found: %w", err)
		}
		folder.Owner = user.ID
        folder.OwnerUsername = user.Username
		if err := tx.Create(&folder).Error; err != nil {
			return fmt.Errorf("failed to create folder: %w", err)
		}
		tx.Model(&user).Association("Folders").Append(&folder)
		return nil
	})
}
func AddPictureToFolder(folderID uint, picture model.Picture, userID uint) error {
    return database.DB.Transaction(func(tx *gorm.DB) error {
        var folder model.Folder
        if err := tx.First(&folder, folderID).Error; err != nil {
            return fmt.Errorf("folder not found: %w", err)
        }
        var user model.User
        if err := tx.First(&user, userID).Error; err != nil {
            return fmt.Errorf("user not found: %w", err)
        }
		if folder.Owner != userID{
			return fmt.Errorf("user does not have permission to do that")
		}
        picture.OwnerID = userID
        picture.FolderID = folder.ID
        if err := tx.Create(&picture).Error; err != nil {
            return fmt.Errorf("failed to create picture: %w", err)
        }
        return nil
    })
}

func AddContributer(folderID uint, userID uint, newUserID uint) error {
    return database.DB.Transaction(func(tx *gorm.DB) error {
        var folder model.Folder
        var newUser model.User
        if err := tx.First(&folder, folderID).Error; err != nil {
            return fmt.Errorf("folder not found: %w", err)
        }
        if err := tx.First(&newUser, newUserID).Error; err != nil {
            return fmt.Errorf("new user not found: %w", err)
        }
        if folder.Owner != userID {
            return fmt.Errorf("user does not have permission to do that")
        }
        if err := tx.Model(&folder).Association("Contributers").Append(&newUser); err != nil {
            return fmt.Errorf("failed to add contributor: %w", err)
        }
        return nil
    })
}
func GetFolder(folderID uint)(model.Folder, error){
	var folder model.Folder
	if err := database.DB.Model(&folder).Preload("Contributers").Preload("Pictures").Find(&folder, folderID).Error; err != nil{
		return model.Folder{},err
	}
	return folder, nil
}