package util

import (
	"fmt"
	"github.com/cartabinaria/polleg/models"
	slogGorm "github.com/orandin/slog-gorm"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var db *gorm.DB = nil

func ConnectDb(ConnStr string, trace bool) error {
	config := &gorm.Config{
		PrepareStmt: true, // optimize raw queries
	}
	if trace {
		config.Logger = slogGorm.New(slogGorm.WithTraceAll())
	}
	var err error
	db, err = gorm.Open(postgres.Open(ConnStr), config)
	if err != nil {
		return fmt.Errorf("failed to open db connection: %w", err)
	}
	return nil
}

func GetDb() *gorm.DB {
	return db
}

func GetUserByID(db *gorm.DB, id uint) (*models.User, error) {
	var user models.User
	if err := db.First(&user, id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func GetOrCreateUserByID(db *gorm.DB, id uint, username string) (*models.User, error) {
	user, err := GetUserByID(db, id)
	if err == nil {
		return user, nil
	} else if err != gorm.ErrRecordNotFound {
		return nil, err
	}

	// Create new user with unique alias
	alias, err := generateUniqueAlias(db)
	if err != nil {
		return nil, err
	}

	user = &models.User{
		ID:       id,
		Username: username,
		Alias:    alias,
	}

	if err := db.Create(user).Error; err != nil {
		return nil, err
	}

	return user, nil
}
