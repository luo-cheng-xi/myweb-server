package po

import "gorm.io/gorm"

type User struct {
	gorm.Model
	Name     string `gorm:"size:32;uniqueIndex:idx_name"`
	Email    string `gorm:"size:320;uniqueIndex:idx_email"`
	Password string `gorm:"size:255"`
	Avatar   string `gorm:"size:255"`
}
