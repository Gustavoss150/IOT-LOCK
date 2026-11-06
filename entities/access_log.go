package entities

import "time"

type AccessLog struct {
	ID         uint      `json:"id" gorm:"primaryKey"`
	CardID     *uint     `json:"card_id" gorm:"index"`
	CardUID    string    `json:"card_uid" gorm:"size:50"`
	Authorized bool      `json:"authorized"`
	AccessedAt time.Time `json:"accessed_at"`

	// Relacionamento
	Card *Card `json:"card,omitempty" gorm:"foreignKey:CardID"`
}

