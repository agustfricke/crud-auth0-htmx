package models

import "gorm.io/gorm"

type Task struct {
	gorm.Model
	Name    string `json:"name"`
    UserId  string `json:"user_id"`
}
