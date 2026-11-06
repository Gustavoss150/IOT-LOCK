package migrations

import (
	"fmt"

	"iot-lock/entities"

	"gorm.io/gorm"
)

func Run(db *gorm.DB) error {
	if db == nil {
		return fmt.Errorf("a conexão com o banco é obrigatória")
	}

	if err := db.AutoMigrate(&entities.User{}, &entities.Card{}, &entities.AccessLog{}); err != nil {
		return fmt.Errorf("migrar tabelas: %w", err)
	}

	statements := []struct {
		name string
		sql  string
	}{
		{
			name: "permitir card_id nulo em access_logs",
			sql:  "ALTER TABLE access_logs ALTER COLUMN card_id DROP NOT NULL",
		},
		{
			name: "sincronizar sequência de users",
			sql:  "SELECT setval(pg_get_serial_sequence('users', 'id'), COALESCE((SELECT MAX(id) FROM users), 1), (SELECT COUNT(*) > 0 FROM users))",
		},
		{
			name: "sincronizar sequência de cards",
			sql:  "SELECT setval(pg_get_serial_sequence('cards', 'id'), COALESCE((SELECT MAX(id) FROM cards), 1), (SELECT COUNT(*) > 0 FROM cards))",
		},
	}

	for _, statement := range statements {
		if err := db.Exec(statement.sql).Error; err != nil {
			return fmt.Errorf("%s: %w", statement.name, err)
		}
	}

	return nil
}
