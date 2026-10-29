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

var _ repoInterfaces.UserRepository = (*userRepository)(nil)

func InitUserRepository(db *gorm.DB) (repoInterfaces.UserRepository, error) {
	if db == nil {
		return nil, errors.New("a conexão com o banco é obrigatória")
	}
	return &userRepository{db: db}, nil
}

func (r *userRepository) Create(user *entities.User) error {
	if user == nil {
		return errors.New("o usuário é obrigatório")
	}
	return r.db.Create(user).Error
}

func (r *userRepository) Save(user *entities.User) error {
	if user == nil {
		return errors.New("o usuário é obrigatório")
	}
	return r.db.Save(user).Error
}

func (r *userRepository) FindByID(id uint) (*entities.User, error) {
	var user entities.User

	if err := r.db.First(&user, id).Error; err != nil {
		return nil, fmt.Errorf("buscar usuário %d: %w", id, err)
	}

	return &user, nil
}

func (r *userRepository) FindAll() ([]*entities.User, error) {
	var users []*entities.User
	if err := r.db.Order("name ASC").Find(&users).Error; err != nil {
		return nil, fmt.Errorf("buscar usuários: %w", err)
	}
	return users, nil
}

func (r *userRepository) DeleteByID(id uint) error {
	result := r.db.Delete(&entities.User{}, "id = ?", id)
	if result.Error != nil {
		return fmt.Errorf("excluir usuário %d: %w", id, result.Error)
	}

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}
