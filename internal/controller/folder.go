package controller

import (
	"fmt"
	"gorm.io/gorm"
	"twilu/internal/model"
)

// FolderController handles operations on folders.
type FolderController struct {
	DB *gorm.DB
}

// NewFolderController creates a new instance of FolderController.
func NewFolderController(db *gorm.DB) *FolderController {
	return &FolderController{DB: db}
}

func (fc *FolderController) CreateFolder(folder model.Folder, userID int) error {
	return fc.DB.Transaction(func(tx *gorm.DB) error {
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
func (fc *FolderController) AddContributer(folderID int, userID int, newUserID int) error {
	return fc.DB.Transaction(func(tx *gorm.DB) error {
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
func (fc *FolderController) GetFolder(folderID int) (model.Folder, error) {
	var folder model.Folder
	if err := fc.DB.Model(&folder).Preload("Contributors").Preload("Items").Find(&folder, folderID).Error; err != nil {
		return model.Folder{}, err
	}
	return folder, nil
}
func (fc *FolderController) DeleteFolder(folderID int, userID int) error {
	return fc.DB.Transaction(func(tx *gorm.DB) error {
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
func (fc *FolderController) GetFeed() ([]model.Folder, error) {
	var folders []model.Folder
	if err := fc.DB.Model(&model.Folder{}).
		Where("Private = ?", false).
		Order("created_at DESC").
		Limit(20).
		Find(&folders).Error; err != nil {
		fmt.Println(folders)
		return []model.Folder{}, err
	}
	return folders, nil
}
