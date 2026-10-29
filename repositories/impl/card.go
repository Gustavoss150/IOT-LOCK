package repoImpl

import (
	"errors"
	"fmt"
	"iot-lock/entities"
	repoInterfaces "iot-lock/repositories/interfaces"

	"gorm.io/gorm"
)

type cardRepository struct {
	db *gorm.DB
}

var _ repoInterfaces.CardRepository = (*cardRepository)(nil)

func InitCardRepository(db *gorm.DB) (repoInterfaces.CardRepository, error) {
	if db == nil {
		return nil, errors.New("a conexão com o banco é obrigatória")
	}
	return &cardRepository{db: db}, nil
}

func (r *cardRepository) Create(card *entities.Card) error {
	if card == nil {
		return errors.New("o cartão é obrigatório")
	}
	return r.db.Create(card).Error
}

// Save atualiza o cartão e, conforme o comportamento do GORM, pode inseri-lo se não existir.
func (r *cardRepository) Save(card *entities.Card) error {
	if card == nil {
		return errors.New("o cartão é obrigatório")
	}
	return r.db.Save(card).Error
}

func (r *cardRepository) FindByID(id uint) (*entities.Card, error) {
	var card entities.Card
	if err := r.db.First(&card, id).Error; err != nil {
		return nil, fmt.Errorf("buscar cartão %d: %w", id, err)
	}
	return &card, nil
}

func (r *cardRepository) FindByUID(uid string) (*entities.Card, error) {
	var card entities.Card
	if err := r.db.Where("card_uid = ?", uid).First(&card).Error; err != nil {
		return nil, fmt.Errorf("buscar cartão pelo UID: %w", err)
	}
	return &card, nil
}

func (r *cardRepository) FindAll() ([]*entities.Card, error) {
	var cards []*entities.Card
	if err := r.db.Order("card_uid ASC").Find(&cards).Error; err != nil {
		return nil, fmt.Errorf("buscar cartões: %w", err)
	}
	return cards, nil
}

func (r *cardRepository) DeleteByID(id uint) error {
	result := r.db.Delete(&entities.Card{}, id)
	if result.Error != nil {
		return fmt.Errorf("excluir cartão %d: %w", id, result.Error)
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
