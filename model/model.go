package model

import (
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Email          string    `gorm:"uniqueIndex;not null"`
	Username       string    `gorm:"not null"`
	Password       string    `gorm:"not null"`
	Folders        []*Folder `gorm:"many2many:user_folders;"`
	ProfilePicture string
	Bio            string
}

type Picture struct {
	gorm.Model
	Name     string
	URL      string `gorm:"not null"`
	FolderID uint
	OwnerID  uint
}

type Folder struct {
	gorm.Model
	Name          string
	Owner         uint
	OwnerUsername string
	Contributors  []*User    `gorm:"many2many:folder_contributors;"`
	Pictures      []*Picture `gorm:"foreignKey:FolderID"`
	Private       bool
	CoverURL      string
}
