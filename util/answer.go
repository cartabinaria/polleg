package util

import (
	"fmt"
	"regexp"
)

func GenerateAnonymousAvatar(alias string) string {
	re := regexp.MustCompile(`^[A-Za-z]+`)
	match := re.FindString(alias)
	if match == "" {
		return ""
	}
	return fmt.Sprintf("https://api.dicebear.com/9.x/thumbs/svg?seed=%s", match)
}
