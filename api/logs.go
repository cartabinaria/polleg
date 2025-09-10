package api

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/cartabinaria/auth/pkg/httputil"
	"github.com/cartabinaria/auth/pkg/middleware"
	"github.com/cartabinaria/polleg/models"
	"github.com/cartabinaria/polleg/util"
)

const SYSTEM_USER_ID = 0

type Log struct {
	Timestamp time.Time `json:"timestamp"`
	Action    string    `json:"action"`    // Created, Updated, Deleted
	ItemType  string    `json:"item_type"` // Answer, Image, ecc.
	ItemID    string    `json:"item_id"`

	UserID        uint   `json:"-"`
	Username      string `json:"username"`
	UserAvatarURL string `json:"user_avatar_url"`
}

func LogsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		httputil.WriteError(w, http.StatusMethodNotAllowed, "invalid method")
		return
	}

	if !middleware.GetAdmin(r) {
		httputil.WriteError(w, http.StatusForbidden, "you are not admin")
		return
	}

	db := util.GetDb()

	// Images
	var images []models.Image
	if err := db.Find(&images).Error; err != nil {
		slog.With("err", err).Error("error while getting images from DB")
		httputil.WriteError(w, http.StatusBadRequest, "could not get logs")
		return
	}

	logs := imagesToLogs(images)

	var users []models.User
	if err := db.Find(&users).Error; err != nil {
		slog.With("err", err).Error("error while getting users from DB")
		httputil.WriteError(w, http.StatusBadRequest, "could not get logs")
		return
	}

	userMap := make(map[uint]string, len(users))
	for _, u := range users {
		userMap[u.ID] = u.Username
	}

	for i := range logs {
		if logs[i].UserID == SYSTEM_USER_ID {
			logs[i].Username = "system"
			logs[i].UserAvatarURL = ""
			continue
		} else if username, ok := userMap[logs[i].UserID]; ok {
			logs[i].Username = username
			logs[i].UserAvatarURL = util.GetPublicAvatarURL(logs[i].UserID)
		}
	}

	httputil.WriteData(w, http.StatusOK, logs)
}

func imagesToLogs(images []models.Image) []Log {
	logs := make([]Log, len(images))
	for _, img := range images {
		logs = append(logs, Log{
			Timestamp: img.CreatedAt,
			Action:    "Created",
			ItemType:  "Image",
			ItemID:    img.ID,

			UserID:        img.UserID,
			Username:      "",
			UserAvatarURL: "",
		})

		if img.DeletedAt.Valid {
			logs = append(logs, Log{
				Timestamp: img.DeletedAt.Time,
				Action:    "Deleted",
				ItemType:  "Image",
				ItemID:    img.ID,

				UserID:        SYSTEM_USER_ID,
				Username:      "",
				UserAvatarURL: "",
			})
		}
	}
	return logs
}
