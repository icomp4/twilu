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
		err := tx.Model(&user).Association("Folders").Append(&folder)
		if err != nil {
			return err
		}
		return nil
	})
}
func AddItemToFolder(folderID int, item model.Item, userID int) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		var folder model.Folder
		userIDUint := uint(userID)
		if err := tx.First(&folder, folderID).Error; err != nil {
			return fmt.Errorf("folder not found: %w", err)
		}
		var user model.User
		if err := tx.First(&user, userID).Error; err != nil {
			return fmt.Errorf("user not found: %w", err)
		}
		if folder.Owner != userIDUint {
			return fmt.Errorf("user does not have permission to do that")
		}
		item.OwnerID = userIDUint
		item.FolderID = folder.ID
		if err := tx.Create(&item).Error; err != nil {
			return fmt.Errorf("failed to create picture: %w", err)
		}
		return nil
	})
}
func AddContributer(folderID int, userID int, newUserID int) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		var folder model.Folder
		var newUser model.User
		folderIDUint := uint(folderID)
		if err := tx.First(&folder, folderID).Error; err != nil {
			return fmt.Errorf("folder not found: %w", err)
		}
		if err := tx.First(&newUser, newUserID).Error; err != nil {
			return fmt.Errorf("new user not found: %w", err)
		}
		if folder.Owner != folderIDUint {
			return fmt.Errorf("user does not have permission to do that")
		}
		if err := tx.Model(&folder).Association("Contributors").Append(&newUser); err != nil {
			return fmt.Errorf("failed to add contributor: %w", err)
		}
		return nil
	})
}
func GetFolder(folderID int) (model.Folder, error) {
	var folder model.Folder
	if err := database.DB.Model(&folder).Preload("Contributors").Preload("Items").Find(&folder, folderID).Error; err != nil {
		return model.Folder{}, err
	}
	return folder, nil
}
func DeleteFolder(folderID int, userID int) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		var user model.User
		var folder model.Folder

		if err := tx.First(&user, userID).Error; err != nil {
			return fmt.Errorf("user not found: %w", err)
		}
		if err := tx.First(&folder, folderID).Error; err != nil {
			return fmt.Errorf("folder not found: %w", err)
		}
		// Check if the user is the owner of the folder
		if userID != int(folder.Owner) {
			return fmt.Errorf("user is not the owner")
		}

		if err := tx.Model(&folder).Association("Contributors").Clear(); err != nil {
			return fmt.Errorf("unable to clear folder contributors: %w", err)
		}

		if err := tx.Model(&user).Association("Folders").Delete(&folder); err != nil {
			return fmt.Errorf("unable to remove folder from user's folders: %w", err)
		}
		if err := tx.Where("Folder_ID = ?", folderID).Unscoped().Delete(&model.Item{}).Error; err != nil {
			return fmt.Errorf("unable to delete items: %w", err)
		}

		if err := tx.Unscoped().Delete(&folder).Error; err != nil {
			return fmt.Errorf("unable to delete folder: %w", err)
		}
		return nil
	})
}
func DeleteItem(folderID int, userID int, itemID int) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		var user model.User
		var folder model.Folder
		var item model.Item
		if err := tx.First(&user, userID).Error; err != nil {
			return fmt.Errorf("user not found: %w", err)
		}
		if err := tx.First(&folder, folderID).Error; err != nil {
			return fmt.Errorf("folder not found: %w", err)
		}
		if err := tx.First(&item, itemID).Error; err != nil {
			return fmt.Errorf("item not found: %w", err)
		}
		if userID != int(item.OwnerID) {
			return fmt.Errorf("user is not the owner")
		}
		if err := tx.Unscoped().Delete(&item).Error; err != nil {
			return fmt.Errorf("unable to delete item: %w", err)
		}
		return nil
	})
}
