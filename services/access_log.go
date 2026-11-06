package services

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"iot-lock/entities"
	repoInterfaces "iot-lock/repositories/interfaces"
)

type AccessLogService struct {
	repo repoInterfaces.AccessLogRepository
}

func NewAccessLogService(repo repoInterfaces.AccessLogRepository) *AccessLogService {
	return &AccessLogService{repo: repo}
}

func (s *AccessLogService) Create(cardID *uint, cardUID string, authorized bool) (*entities.AccessLog, error) {
	uid := strings.TrimSpace(cardUID)
	if (cardID == nil || *cardID == 0) && uid == "" {
		return nil, errors.New("o ID ou UID do cartão é obrigatório")
	}

	var normalizedCardID *uint
	if cardID != nil && *cardID > 0 {
		normalizedCardID = cardID
	}

	log := &entities.AccessLog{
		CardID:     normalizedCardID,
		CardUID:    uid,
		Authorized: authorized,
		AccessedAt: time.Now(),
	}

	if err := s.repo.Create(log); err != nil {
		return nil, fmt.Errorf("registrar acesso: %w", err)
	}

	return log, nil
}

func (s *AccessLogService) GetByID(id uint) (*entities.AccessLog, error) {
	if id == 0 {
		return nil, errors.New("o ID do registro é obrigatório")
	}
	return s.repo.FindByID(id)
}

func (s *AccessLogService) ListRecent(limit, offset int) ([]*entities.AccessLog, error) {
	if limit <= 0 {
		limit = 20
	}
	if offset < 0 {
		return nil, errors.New("o deslocamento não pode ser negativo")
	}
	return s.repo.FindRecent(limit, offset)
}

func (s *AccessLogService) ListByCardID(cardID uint, limit, offset int) ([]*entities.AccessLog, error) {
	if cardID == 0 {
		return nil, errors.New("o ID do cartão é obrigatório")
	}
	if limit <= 0 {
		limit = 20
	}
	if offset < 0 {
		return nil, errors.New("o deslocamento não pode ser negativo")
	}
	return s.repo.FindByCardID(cardID, limit, offset)
}
