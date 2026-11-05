package services

import (
	"errors"
	"fmt"
	"strings"

	"iot-lock/entities"
	repoInterfaces "iot-lock/repositories/interfaces"
)

type CardService struct {
	repo repoInterfaces.CardRepository
}

func NewCardService(repo repoInterfaces.CardRepository) *CardService {
	return &CardService{repo: repo}
}

func (s *CardService) Create(cardUID string, userID uint) (*entities.Card, error) {
	if strings.TrimSpace(cardUID) == "" {
		return nil, errors.New("o UID do cartão é obrigatório")
	}
	if userID == 0 {
		return nil, errors.New("o ID do usuário é obrigatório")
	}

	card := &entities.Card{
		CardUID: strings.TrimSpace(cardUID),
		UserID:  userID,
		Active:  true,
	}

	if err := s.repo.Create(card); err != nil {
		return nil, fmt.Errorf("criar cartão: %w", err)
	}

	return card, nil
}

func (s *CardService) Update(card *entities.Card) (*entities.Card, error) {
	if card == nil {
		return nil, errors.New("o cartão é obrigatório")
	}
	if strings.TrimSpace(card.CardUID) == "" {
		return nil, errors.New("o UID do cartão é obrigatório")
	}
	if card.UserID == 0 {
		return nil, errors.New("o ID do usuário é obrigatório")
	}

	if err := s.repo.Save(card); err != nil {
		return nil, fmt.Errorf("atualizar cartão: %w", err)
	}

	return card, nil
}

func (s *CardService) UpdateStatus(id uint, active bool) (*entities.Card, error) {
	if id == 0 {
		return nil, errors.New("o ID do cartão é obrigatório")
	}

	card, err := s.repo.FindByID(id)
	if err != nil {
		return nil, fmt.Errorf("buscar cartão para atualizar status: %w", err)
	}

	card.Active = active
	if err := s.repo.Save(card); err != nil {
		return nil, fmt.Errorf("atualizar status do cartão: %w", err)
	}

	return card, nil
}

func (s *CardService) GetByID(id uint) (*entities.Card, error) {
	if id == 0 {
		return nil, errors.New("o ID do cartão é obrigatório")
	}
	return s.repo.FindByID(id)
}

func (s *CardService) GetByUID(uid string) (*entities.Card, error) {
	if strings.TrimSpace(uid) == "" {
		return nil, errors.New("o UID do cartão é obrigatório")
	}
	return s.repo.FindByUID(uid)
}

func (s *CardService) List() ([]*entities.Card, error) {
	return s.repo.FindAll()
}

func (s *CardService) Delete(id uint) error {
	if id == 0 {
		return errors.New("o ID do cartão é obrigatório")
	}
	return s.repo.DeleteByID(id)
}
