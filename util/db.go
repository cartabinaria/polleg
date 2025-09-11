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

func SaveNewReport(db *gorm.DB, answerID uint, cause string, userID uint) error {
	report := models.Report{
		AnswerID: answerID,
		Cause:    cause,
		UserID:   userID,
	}
	if err := db.Create(&report).Error; err != nil {
		return err
	}
	return nil
}

func GetAllReports(db *gorm.DB) ([]models.Report, error) {
	var reports []models.Report
	if err := db.Find(&reports).Error; err != nil {
		return nil, err
	}
	return reports, nil
}

func GetBannedUsers(db *gorm.DB) ([]models.User, error) {
	var users []models.User
	if err := db.Where("banned = ?", true).Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

func BanUnbanUser(db *gorm.DB, userID uint, ban bool) error {
	var user models.User
	if err := db.First(&user, userID).Error; err != nil {
		return err
	}

	if ban {
		now := time.Now()
		user.Banned = true
		user.BannedAt = &now
	} else {
		user.Banned = false
		user.BannedAt = nil
	}

	if err := db.Save(&user).Error; err != nil {
		return err
	}
	return nil
}
