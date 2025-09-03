package util

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/cartabinaria/polleg/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gorm_logger "gorm.io/gorm/logger"
)

var db *gorm.DB = nil

func ConnectDb(ConnStr string) error {
	config := &gorm.Config{
		PrepareStmt: true, // optimize raw queries
		Logger: gorm_logger.New(
			log.New(os.Stdout, "\r\n", log.LstdFlags),
			gorm_logger.Config{
				SlowThreshold:             time.Second,
				LogLevel:                  gorm_logger.Error,
				IgnoreRecordNotFoundError: true,
				Colorful:                  true,
			},
		),
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

func CreateImage(db *gorm.DB, id string, userID uint, size uint) (*models.Image, error) {
	image := models.Image{
		ID:     id,
		UserID: userID,
		Size:   size,
	}
	if err := db.Create(&image).Error; err != nil {
		return nil, err
	}

	return &image, nil
}

func GetTotalSizeOfImagesByUser(db *gorm.DB, userID uint) (uint64, error) {
	var totalSize uint64
	err := db.Model(&models.Image{}).Where("user_id = ?", userID).Select("COALESCE(SUM(size), 0)").Scan(&totalSize).Error
	if err != nil {
		return 0, err
	}
	return totalSize, nil
}

func GetNumberOfImagesByUser(db *gorm.DB, userID uint) (int64, error) {
	var count int64
	err := db.Model(&models.Image{}).Where("user_id = ?", userID).Count(&count).Error
	if err != nil {
		return 0, err
	}
	return count, nil
}
