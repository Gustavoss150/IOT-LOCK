package config

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"os"
	"strings"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var supabasePoolerRegions = []string{
	"aws-0-us-east-2",
	"aws-1-us-east-2",
	"aws-0-sa-east-1",
	"aws-1-sa-east-1",
	"aws-0-us-east-1",
	"aws-1-us-east-1",
}

func resolveCandidateDatabaseURLs() ([]string, error) {
	if databaseURL := strings.TrimSpace(os.Getenv("DATABASE_URL")); databaseURL != "" {
		return []string{databaseURL}, nil
	}

	supabaseURL := strings.TrimSpace(os.Getenv("SUPABASE_URL"))
	dbPassword := strings.TrimSpace(os.Getenv("DATABASE_PASSWORD"))
	if supabaseURL != "" && dbPassword != "" {
		parsed, err := url.Parse(supabaseURL)
		if err == nil && parsed.Host != "" {
			projectRef := strings.Split(parsed.Host, ".")[0]
			if projectRef != "" {
				escapedPass := url.QueryEscape(dbPassword)
				var candidates []string

				if region := strings.TrimSpace(os.Getenv("SUPABASE_REGION")); region != "" {
					host := fmt.Sprintf("%s.pooler.supabase.com", region)
					candidates = append(candidates, fmt.Sprintf(
						"postgresql://postgres.%s:%s@%s:5432/postgres?sslmode=require",
						projectRef, escapedPass, host,
					))
				}

				for _, region := range supabasePoolerRegions {
					host := fmt.Sprintf("%s.pooler.supabase.com", region)
					candidates = append(candidates, fmt.Sprintf(
						"postgresql://postgres.%s:%s@%s:5432/postgres?sslmode=require",
						projectRef, escapedPass, host,
					))
				}

				// Fallback direto (IPv6)
				candidates = append(candidates, fmt.Sprintf(
					"postgresql://postgres:%s@db.%s.supabase.co:5432/postgres?sslmode=require",
					escapedPass, projectRef,
				))
				return candidates, nil
			}
		}
	}

	return nil, fmt.Errorf("defina DATABASE_URL ou (SUPABASE_URL e DATABASE_PASSWORD) no arquivo .env")
}

func openAndPing(dsn string) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	sqlDB.SetMaxOpenConns(10)
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetConnMaxLifetime(30 * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()
	if err := sqlDB.PingContext(ctx); err != nil {
		_ = sqlDB.Close()
		return nil, err
	}

	db.Logger = logger.Default.LogMode(logger.Info)
	return db, nil
}

func ConnectDatabase() (*gorm.DB, error) {
	candidates, err := resolveCandidateDatabaseURLs()
	if err != nil {
		return nil, err
	}

	var lastErr error
	for _, dsn := range candidates {
		db, err := openAndPing(dsn)
		if err == nil {
			log.Println("✅ Conectado ao PostgreSQL do Supabase!")
			return db, nil
		}
		lastErr = err
		errLower := strings.ToLower(err.Error())
		if !strings.Contains(errLower, "not found") &&
			!strings.Contains(errLower, "hostname resolving error") &&
			!strings.Contains(errLower, "no such host") {
			break
		}
	}

	return nil, fmt.Errorf("conectar ao PostgreSQL do Supabase: %w", lastErr)
}

