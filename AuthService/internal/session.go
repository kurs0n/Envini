package internal

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Session struct {
	SessionID             uuid.UUID `gorm:"type:uuid;primaryKey"`
	GithubUserID          int64     `gorm:"uniqueIndex"`
	UserLogin             string    `gorm:"size:255;not null"` // GitHub user login
	AccessToken           string
	RefreshToken          string
	ExpiresAt             time.Time
	RefreshTokenExpiresAt time.Time
	CreatedAt             time.Time
}

type SessionStore struct {
	DB *gorm.DB
}

func NewSessionStore(dsn string) (*SessionStore, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	if err := db.AutoMigrate(&Session{}); err != nil {
		return nil, err
	}
	return &SessionStore{DB: db}, nil
}

func (s *SessionStore) UpsertByGithubUserID(ctx context.Context, session *Session) error {
	return s.DB.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "github_user_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"session_id", "access_token", "refresh_token", "expires_at", "refresh_token_expires_at", "created_at"}),
	}).Create(session).Error
}

func (s *SessionStore) GetBySessionID(ctx context.Context, sessionID uuid.UUID) (*Session, error) {
	var session Session
	if err := s.DB.WithContext(ctx).First(&session, "session_id = ?", sessionID).Error; err != nil {
		return nil, err
	}
	return &session, nil
}

func (s *SessionStore) Delete(ctx context.Context, sessionID uuid.UUID) error {
	return s.DB.WithContext(ctx).Delete(&Session{}, "session_id = ?", sessionID).Error
}
