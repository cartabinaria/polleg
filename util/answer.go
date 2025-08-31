package util

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"regexp"
	"strconv"

	"github.com/cartabinaria/polleg/models"
	"gorm.io/gorm"
)

var names = []string{"Wyatt", "Vivian", "Maria", "Alexander", "Luis", "Aidan", "Mason", "Aiden", "Mackenzie", "Adrian", "Oliver", "Andrea", "Amaya", "Nolan", "Riley", "Robert", "Ryker", "Sara", "Ryan", "Sawyer"}

func GenerateAnonymousAvatar(alias string) string {
	re := regexp.MustCompile(`^[A-Za-z]+`)
	match := re.FindString(alias)
	if match == "" {
		return ""
	}
	return fmt.Sprintf("https://api.dicebear.com/9.x/thumbs/svg?seed=%s", match)
}

func generateUniqueAlias(db *gorm.DB) (string, error) {
	// Generate cryptographically secure random index
	nameIndex, err := rand.Int(rand.Reader, big.NewInt(int64(len(names))))
	if err != nil {
		return "", fmt.Errorf("failed to generate random name index: %w", err)
	}

	name := names[nameIndex.Int64()]
	var lastUser models.User
	pattern := fmt.Sprintf("%s_%%", name)
	result := db.Where("alias LIKE ?", pattern).Order("created_at DESC").First(&lastUser)

	if result.Error != nil {
		if result.Error != gorm.ErrRecordNotFound {
			return "", fmt.Errorf("failed to get last user: %w", result.Error)
		} else {
			return fmt.Sprintf("%s_1", name), nil
		}
	}

	nextNum := 1

	re := regexp.MustCompile(fmt.Sprintf(`^%s_(\d+)$`, name))
	if matches := re.FindStringSubmatch(lastUser.Alias); len(matches) == 2 {
		if n, err := strconv.Atoi(matches[1]); err == nil {
			nextNum = n + 1
		}
	}
	return fmt.Sprintf("%s_%d", name, nextNum), nil
}
