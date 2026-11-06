package config

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"os"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func ConnectDatabase() (*gorm.DB, error) {
	databaseURL, err := supabaseDatabaseURL()
	if err != nil {
		return nil, err
	}

	db, err := gorm.Open(
		postgres.New(postgres.Config{
			DSN:                  databaseURL,
			PreferSimpleProtocol: true,
		}),
		&gorm.Config{},
	)
	if err != nil {
		return nil, fmt.Errorf("conectar ao PostgreSQL do Supabase: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("obter conexão SQL: %w", err)
	}

	sqlDB.SetMaxOpenConns(10)
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetConnMaxLifetime(30 * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := sqlDB.PingContext(ctx); err != nil {
		_ = sqlDB.Close()
		return nil, fmt.Errorf("verificar conexão com o PostgreSQL do Supabase: %w", err)
	}

	return db, nil
}

func supabaseDatabaseURL() (string, error) {
	host := os.Getenv("SUPABASE_DB_HOST")
	port := os.Getenv("SUPABASE_DB_PORT")
	databaseName := os.Getenv("SUPABASE_DB_NAME")
	databaseUser := os.Getenv("SUPABASE_DB_USER")
	databasePassword := os.Getenv("DATABASE_PASSWORD")

	if host == "" || port == "" || databaseName == "" ||
		databaseUser == "" || databasePassword == "" {
		return "", fmt.Errorf(
			"defina SUPABASE_DB_HOST, SUPABASE_DB_PORT, SUPABASE_DB_NAME, SUPABASE_DB_USER e DATABASE_PASSWORD",
		)
	}

	connectionURL := url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(databaseUser, databasePassword),
		Host:   net.JoinHostPort(host, port),
		Path:   "/" + databaseName,
		RawQuery: url.Values{
			"sslmode": {"require"},
		}.Encode(),
	}

	return connectionURL.String(), nil
}
