package storage

import (
	"context"
	"database/sql"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type PostgresStorage struct {
	db *gorm.DB
}

func SetupPostgresStorage(dns string) error {
	db, err := gorm.Open(postgres.Open(dns), &gorm.Config{})
	if err != nil {
		return err
	}
	db.AutoMigrate(&Session{})
	return nil
}

func NewPostgresStorage(db *sql.DB) Storage {
	gormDB, err := gorm.Open(postgres.New(postgres.Config{Conn: db}), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	return &PostgresStorage{
		db: gormDB,
	}
}

func NewPostgresStorageWithDNS(dns string) Storage {
	gormDB, err := gorm.Open(postgres.Open(dns), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	return &PostgresStorage{
		db: gormDB,
	}
}

func NewPostgresStorageWithGormDB(gormDB *gorm.DB) Storage {
	return &PostgresStorage{
		db: gormDB,
	}
}

func (s *PostgresStorage) AddNewJourney(ctx context.Context, journey *Journey) error {
	return s.db.Create(journey).Error
}

func (s *PostgresStorage) SaveSession(ctx context.Context, session *Session) error {
	return s.db.Create(session).Error
}

func (s *PostgresStorage) FetchSession(ctx context.Context, id string) (*Session, error) {
	var session Session
	if err := s.db.Where("id = ?", id).First(&session).Error; err != nil {
		return nil, err
	}
	return &session, nil
}

func (s *PostgresStorage) UpdateSession(ctx context.Context, session *Session) error {
	return s.db.Save(session).Error
}

func (s *PostgresStorage) DeleteSession(ctx context.Context, id string) error {
	return s.db.Delete(&Session{}, "id = ?", id).Error
}
