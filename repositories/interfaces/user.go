package repoInterfaces

import "iot-lock/entities"

type UserRepository interface {
	Create(user *entities.User) error
	Save(user *entities.User) error
	FindByID(id uint) (*entities.User, error)
	FindAll() ([]*entities.User, error)
	DeleteByID(id uint) error
}
