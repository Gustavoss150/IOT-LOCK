package repoImpl

import (
	"errors"
	"fmt"
	"iot-lock/entities"
	repoInterfaces "iot-lock/repositories/interfaces"

	"gorm.io/gorm"
)

type accessLogRepository struct {
	db *gorm.DB
}

var _ repoInterfaces.AccessLogRepository = (*accessLogRepository)(nil)

func InitAccessLogRepository(db *gorm.DB) (repoInterfaces.AccessLogRepository, error) {
	if db == nil {
		return nil, errors.New("a conexão com o banco é obrigatória")
	}
	return &accessLogRepository{db: db}, nil
}

func (r *accessLogRepository) Create(log *entities.AccessLog) error {
	if log == nil {
		return errors.New("o registro de acesso é obrigatório")
	}
	return r.db.Create(log).Error
}

func (r *accessLogRepository) FindByID(id uint) (*entities.AccessLog, error) {
	var log entities.AccessLog
	if err := r.db.First(&log, id).Error; err != nil {
		return nil, fmt.Errorf("buscar registro de acesso %d: %w", id, err)
	}
	return &log, nil
}

func (r *accessLogRepository) FindRecent(limit, offset int) ([]*entities.AccessLog, error) {
	if err := validatePagination(limit, offset); err != nil {
		return nil, err
	}

	var logs []*entities.AccessLog
	if err := r.db.Order("accessed_at DESC").Order("id DESC").Limit(limit).Offset(offset).Find(&logs).Error; err != nil {
		return nil, fmt.Errorf("buscar registros de acesso recentes: %w", err)
	}
	return logs, nil
}

func (r *accessLogRepository) FindByCardID(cardID uint, limit, offset int) ([]*entities.AccessLog, error) {
	if err := validatePagination(limit, offset); err != nil {
		return nil, err
	}

	var logs []*entities.AccessLog
	if err := r.db.Where("card_id = ?", cardID).
		Order("accessed_at DESC").
		Order("id DESC").
		Limit(limit).
		Offset(offset).
		Find(&logs).Error; err != nil {
		return nil, fmt.Errorf("buscar registros de acesso do cartão %d: %w", cardID, err)
	}
	return logs, nil
}

func validatePagination(limit, offset int) error {
	if limit <= 0 {
		return errors.New("o limite deve ser maior que zero")
	}
	if offset < 0 {
		return errors.New("o deslocamento não pode ser negativo")
	}
	return nil
}
