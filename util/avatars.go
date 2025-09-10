package util

import "fmt"

func GetPublicAvatarURL(userID uint) string {
	return fmt.Sprintf("https://avatars.githubusercontent.com/u/%d?v=4", userID)
}

func GenerateAnonymousAvatar(alias string) string {
	return fmt.Sprintf("https://api.dicebear.com/9.x/thumbs/svg?seed=%s", alias)
}
