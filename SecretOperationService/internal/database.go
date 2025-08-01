package internal

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Database models using GORM

type Repository struct {
	ID          uint      `gorm:"primaryKey;autoIncrement"`
	OwnerLogin  string    `gorm:"size:255;not null"`
	RepoName    string    `gorm:"size:255;not null"`
	RepoID      int64     `gorm:"not null"`
	FullName    string    `gorm:"size:500;not null"`
	HTMLURL     string    `gorm:"size:1000;not null"`
	Description string    `gorm:"type:text"`
	IsPrivate   bool      `gorm:"default:false"`
	CreatedAt   time.Time `gorm:"autoCreateTime"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime"`
	Secrets     []Secret  `gorm:"foreignKey:RepoID;constraint:OnDelete:CASCADE"`

	// Composite unique constraint
	UniqueConstraint string `gorm:"uniqueIndex:idx_owner_repo;size:1"`
}

func (Repository) TableName() string {
	return "repositories"
}

type Secret struct {
	ID           uint      `gorm:"primaryKey;autoIncrement"`
	RepoID       uint      `gorm:"not null"`
	Version      int       `gorm:"not null"`
	Tag          string    `gorm:"size:255"`
	EnvData      string    `gorm:"type:text;not null"` // Changed from JSONB to TEXT for encrypted data
	Checksum     string    `gorm:"size:64;not null"`
	UploadedBy   string    `gorm:"size:255;not null"`
	CreatedAt    time.Time `gorm:"autoCreateTime"`
	IsEncrypted  bool      `gorm:"default:false"`
	EncryptedKey string    `gorm:"size:255"` // Encrypted per-secret key

	// Composite unique constraint
	UniqueConstraint string `gorm:"uniqueIndex:idx_repo_version;size:1"`
}

func (Secret) TableName() string {
	return "secrets"
}

type AuditLog struct {
	ID           uint      `gorm:"primaryKey;autoIncrement"`
	Operation    string    `gorm:"size:50;not null"`
	RepoID       *uint     `gorm:"index"`
	SecretID     *uint     `gorm:"index"`
	UserLogin    string    `gorm:"size:255;not null"`
	IPAddress    string    `gorm:"size:45"` // IPv6 max length
	UserAgent    string    `gorm:"type:text"`
	Success      bool      `gorm:"not null"`
	ErrorMessage string    `gorm:"type:text"`
	CreatedAt    time.Time `gorm:"autoCreateTime;index"`
}

func (AuditLog) TableName() string {
	return "audit_logs"
}

// Database connection
var DB *gorm.DB

// Encryption functions
func generateSecretKey() ([]byte, error) {
	key := make([]byte, 32) // AES-256
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		return nil, err
	}
	return key, nil
}

func getMasterKey() ([]byte, error) {
	masterKey := os.Getenv("MASTER_ENCRYPTION_KEY")
	if masterKey == "" {
		return nil, fmt.Errorf("MASTER_ENCRYPTION_KEY not set")
	}

	// Decode base64 master key
	key, err := base64.StdEncoding.DecodeString(masterKey)
	if err != nil {
		return nil, fmt.Errorf("invalid master key format: %v", err)
	}

	if len(key) != 32 {
		return nil, fmt.Errorf("master key must be 32 bytes (AES-256)")
	}

	return key, nil
}

func encryptData(data []byte, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	return gcm.Seal(nonce, nonce, data, nil), nil
}

func decryptData(encryptedData []byte, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(encryptedData) < nonceSize {
		return nil, fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertext := encryptedData[:nonceSize], encryptedData[nonceSize:]
	return gcm.Open(nil, nonce, ciphertext, nil)
}

func encryptSecretKey(secretKey []byte) ([]byte, error) {
	masterKey, err := getMasterKey()
	if err != nil {
		return nil, err
	}
	return encryptData(secretKey, masterKey)
}

func decryptSecretKey(encryptedSecretKey []byte) ([]byte, error) {
	masterKey, err := getMasterKey()
	if err != nil {
		return nil, err
	}
	return decryptData(encryptedSecretKey, masterKey)
}

// InitDatabase initializes the database connection and runs migrations
func InitDatabase() error {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=UTC",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_PORT"),
	)

	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return fmt.Errorf("failed to connect to database: %v", err)
	}

	// Auto migrate the schema - GORM will handle the order automatically
	err = DB.AutoMigrate(&Repository{}, &Secret{}, &AuditLog{})
	if err != nil {
		return fmt.Errorf("failed to migrate database: %v", err)
	}

	log.Println("Database connected and migrated successfully")
	return nil
}

// Database operations

// GetOrCreateRepository gets an existing repository or creates a new one
func GetOrCreateRepository(ownerLogin, repoName string, repoID int64, fullName, htmlURL, description string, isPrivate bool) (*Repository, error) {
	var repo Repository

	// Try to find existing repository
	result := DB.Where("owner_login = ? AND repo_name = ?", ownerLogin, repoName).First(&repo)
	if result.Error == nil {
		// Repository exists, update if needed
		repo.RepoID = repoID
		repo.FullName = fullName
		repo.HTMLURL = htmlURL
		repo.Description = description
		repo.IsPrivate = isPrivate
		DB.Save(&repo)
		return &repo, nil
	}

	// Create new repository
	repo = Repository{
		OwnerLogin:  ownerLogin,
		RepoName:    repoName,
		RepoID:      repoID,
		FullName:    fullName,
		HTMLURL:     htmlURL,
		Description: description,
		IsPrivate:   isPrivate,
	}

	result = DB.Create(&repo)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to create repository: %v", result.Error)
	}

	return &repo, nil
}

// GetNextVersion gets the next version number for a repository
func GetNextVersion(repoID uint) (int, error) {
	var maxVersion int
	result := DB.Model(&Secret{}).
		Where("repo_id = ?", repoID).
		Select("COALESCE(MAX(version), 0)").
		Scan(&maxVersion)

	if result.Error != nil {
		return 0, fmt.Errorf("failed to get next version: %v", result.Error)
	}

	return maxVersion + 1, nil
}

// CreateSecret creates a new secret version with optional encryption
func CreateSecret(repoID uint, version int, tag, envData, checksum, uploadedBy string, encrypt bool) (*Secret, error) {
	var encryptedKey string
	var finalEnvData string

	if encrypt {
		// Generate a unique key for this secret
		secretKey, err := generateSecretKey()
		if err != nil {
			return nil, fmt.Errorf("failed to generate secret key: %v", err)
		}

		// Encrypt the env data
		encryptedData, err := encryptData([]byte(envData), secretKey)
		if err != nil {
			return nil, fmt.Errorf("failed to encrypt data: %v", err)
		}

		// Encrypt the secret key with master key
		encryptedSecretKey, err := encryptSecretKey(secretKey)
		if err != nil {
			return nil, fmt.Errorf("failed to encrypt secret key: %v", err)
		}

		// Store encrypted data and key
		finalEnvData = base64.StdEncoding.EncodeToString(encryptedData)
		encryptedKey = base64.StdEncoding.EncodeToString(encryptedSecretKey)
	} else {
		finalEnvData = envData
	}

	secret := &Secret{
		RepoID:       repoID,
		Version:      version,
		Tag:          tag,
		EnvData:      finalEnvData,
		Checksum:     checksum,
		UploadedBy:   uploadedBy,
		IsEncrypted:  encrypt,
		EncryptedKey: encryptedKey,
	}

	result := DB.Create(secret)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to create secret: %v", result.Error)
	}

	return secret, nil
}

// GetSecretByVersion gets a specific version of a secret
func GetSecretByVersion(repoID uint, version int) (*Secret, error) {
	var secret Secret
	result := DB.Where("repo_id = ? AND version = ?", repoID, version).First(&secret)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to get secret: %v", result.Error)
	}
	return &secret, nil
}

// GetSecretByTag gets a secret by tag (returns the latest version with that tag)
func GetSecretByTag(repoID uint, tag string) (*Secret, error) {
	var secret Secret
	result := DB.Where("repo_id = ? AND tag = ?", repoID, tag).
		Order("version DESC").
		First(&secret)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to get secret by tag: %v", result.Error)
	}
	return &secret, nil
}

// GetLatestSecret gets the latest version of a secret
func GetLatestSecret(repoID uint) (*Secret, error) {
	var secret Secret
	result := DB.Where("repo_id = ?", repoID).
		Order("version DESC").
		First(&secret)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to get latest secret: %v", result.Error)
	}
	return &secret, nil
}

// ListSecretVersions gets all versions of secrets for a repository
func ListSecretVersions(repoID uint) ([]Secret, error) {
	var secrets []Secret
	result := DB.Where("repo_id = ?", repoID).
		Order("version DESC").
		Find(&secrets)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to list secret versions: %v", result.Error)
	}
	return secrets, nil
}

// DecryptSecretData decrypts the secret data if it's encrypted
func DecryptSecretData(secret *Secret) (string, error) {
	if !secret.IsEncrypted {
		return secret.EnvData, nil
	}

	// Decode the encrypted secret key
	encryptedSecretKeyBytes, err := base64.StdEncoding.DecodeString(secret.EncryptedKey)
	if err != nil {
		return "", fmt.Errorf("failed to decode encrypted secret key: %v", err)
	}

	// Decrypt the secret key
	secretKey, err := decryptSecretKey(encryptedSecretKeyBytes)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt secret key: %v", err)
	}

	// Decode the encrypted data
	encryptedDataBytes, err := base64.StdEncoding.DecodeString(secret.EnvData)
	if err != nil {
		return "", fmt.Errorf("failed to decode encrypted data: %v", err)
	}

	// Decrypt the data
	decryptedData, err := decryptData(encryptedDataBytes, secretKey)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt data: %v", err)
	}

	return string(decryptedData), nil
}

// DeleteSecret deletes a specific version of a secret
func DeleteSecret(repoID uint, version int) error {
	result := DB.Where("repo_id = ? AND version = ?", repoID, version).Delete(&Secret{})
	if result.Error != nil {
		return fmt.Errorf("failed to delete secret: %v", result.Error)
	}
	return nil
}

// DeleteAllSecrets deletes all versions of secrets for a repository
func DeleteAllSecrets(repoID uint) error {
	result := DB.Where("repo_id = ?", repoID).Delete(&Secret{})
	if result.Error != nil {
		return fmt.Errorf("failed to delete all secrets: %v", result.Error)
	}
	return nil
}

// LogAuditEvent logs an audit event
func LogAuditEvent(operation string, repoID *uint, secretID *uint, userLogin, ipAddress, userAgent string, success bool, errorMessage string) error {
	auditLog := &AuditLog{
		Operation:    operation,
		RepoID:       repoID,
		SecretID:     secretID,
		UserLogin:    userLogin,
		IPAddress:    ipAddress,
		UserAgent:    userAgent,
		Success:      success,
		ErrorMessage: errorMessage,
	}

	result := DB.Create(auditLog)
	if result.Error != nil {
		return fmt.Errorf("failed to log audit event: %v", result.Error)
	}

	return nil
}
