package models

import "gorm.io/gorm"

type User struct {
	gorm.Model
	Name     string `json:name binding:"required"`
	Email    string `json:email gorm:"unique" binding:"required, email"`
	Password string `json:password binding:"required"`
}