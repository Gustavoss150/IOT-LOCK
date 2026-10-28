package repoImpl

import (
	"errors"
	"fmt"
	"iot-lock/entities"
	repoInterfaces "iot-lock/repositories/interfaces"

	"gorm.io/gorm"
)

type userRepository struct {
	db *gorm.DB
}

func InitUserRepository(db *gorm.DB) (repoInterfaces.UserRepository, error) {
	if db == nil {
		return nil, errors.New("a conexão com o banco é obrigatória")
	}
	return &userRepository{db: db}, nil
}

func (r *userRepository) Create(user *entities.User) error {
	return r.db.Create(user).Error
}

func (r *userRepository) FindByID(id uint) (*entities.User, error) {
	var user entities.User

	if err := r.db.First(&user, id).Error; err != nil {
		return nil, fmt.Errorf("buscar usuário %d: %w", id, err)
	}

	return &user, nil
}
