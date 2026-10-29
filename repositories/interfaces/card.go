package repoInterfaces

import "iot-lock/entities"

type CardRepository interface {
	Create(card *entities.Card) error
	Save(card *entities.Card) error
	FindByID(id uint) (*entities.Card, error)
	FindByUID(uid string) (*entities.Card, error)
	FindAll() ([]*entities.Card, error)
	DeleteByID(id uint) error
}
