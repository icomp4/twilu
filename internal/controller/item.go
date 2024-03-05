package controller

import (
	"fmt"
	"gorm.io/gorm"
	"twilu/internal/model"
)

// ItemController handles operations on folders.
type ItemController struct {
	DB *gorm.DB
}

// NewItemController creates a new instance of ItemController.
func NewItemController(db *gorm.DB) *ItemController {
	return &ItemController{DB: db}
}

func (ic *ItemController) AddItemToFolder(folderID int, item model.Item, userID int) error {
	return ic.DB.Transaction(func(tx *gorm.DB) error {
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
func (ic *ItemController) DeleteItem(folderID int, userID int, itemID int) error {
	return ic.DB.Transaction(func(tx *gorm.DB) error {
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
