package repoInterfaces

import "iot-lock/entities"

type UserRepository interface {
	Create(user *entities.User) error
}
