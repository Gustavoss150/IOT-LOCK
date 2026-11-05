package services

import (
	"errors"
	"fmt"
	"strings"

	"iot-lock/entities"
	repoInterfaces "iot-lock/repositories/interfaces"
)

type UserService struct {
	repo repoInterfaces.UserRepository
}

func NewUserService(repo repoInterfaces.UserRepository) *UserService {
	return &UserService{repo: repo}
}

func (s *UserService) Create(name string, cpf int64) (*entities.User, error) {
	if strings.TrimSpace(name) == "" {
		return nil, errors.New("o nome do usuário é obrigatório")
	}
	if cpf == 0 {
		return nil, errors.New("o CPF do usuário é obrigatório")
	}

	user := &entities.User{
		Name: strings.TrimSpace(name),
		CPF:  cpf,
	}

	if err := s.repo.Create(user); err != nil {
		return nil, fmt.Errorf("criar usuário: %w", err)
	}

	return user, nil
}

func (s *UserService) Update(user *entities.User) (*entities.User, error) {
	if user == nil {
		return nil, errors.New("o usuário é obrigatório")
	}
	if strings.TrimSpace(user.Name) == "" {
		return nil, errors.New("o nome do usuário é obrigatório")
	}
	if user.CPF == 0 {
		return nil, errors.New("o CPF do usuário é obrigatório")
	}

	if err := s.repo.Save(user); err != nil {
		return nil, fmt.Errorf("atualizar usuário: %w", err)
	}

	return user, nil
}

func (s *UserService) GetByID(id uint) (*entities.User, error) {
	if id == 0 {
		return nil, errors.New("o ID do usuário é obrigatório")
	}
	return s.repo.FindByID(id)
}

func (s *UserService) List() ([]*entities.User, error) {
	return s.repo.FindAll()
}

func (s *UserService) Delete(id uint) error {
	if id == 0 {
		return errors.New("o ID do usuário é obrigatório")
	}
	return s.repo.DeleteByID(id)
}
