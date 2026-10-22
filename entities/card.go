package entities

import "time"

type Card struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	CardUID   string    `json:"card_uid" gorm:"uniqueIndex;size:50;not null"`
	UserID    uint      `json:"user_id" gorm:"index"`
	Active    bool      `json:"active" gorm:"default:true"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// 	relacionamnetos
	User User `json:"user" gorm:"foreignKey:UserID"`
}
