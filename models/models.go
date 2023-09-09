package models

import "gorm.io/gorm"

type User struct {
	gorm.Model
	Nickname    string `gorm:"not null"`
	Picture     string 
	Sub         string 
	Tasks       []Task 
}

type Task struct {
	gorm.Model
	Name        string  `gorm:"not null"`
	UserID      uint    
    User        User    `gorm:"foreignKey:UserID"`
}
