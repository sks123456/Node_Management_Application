package models

type User struct {
	ID       uint   `gorm:"primaryKey"`
	Name     string `gorm:"size:100;not null"`
	Email    string `gorm:"size:100;unique;not null"`
	Password string `gorm:"size:255;not null"` // Store hashed passwords
}
