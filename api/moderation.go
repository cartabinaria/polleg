package api

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/cartabinaria/auth/pkg/httputil"
	"github.com/cartabinaria/auth/pkg/middleware"
	"github.com/cartabinaria/polleg/models"
	"github.com/cartabinaria/polleg/util"
	"github.com/kataras/muxie"
)

type ReportRequest struct {
	Cause string `json:"cause"`
}

type Report struct {
	ID        uint      `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	AnswerID      uint   `json:"answer_id"`
	Cause         string `json:"cause"`
	Username      string `json:"username"`
	UserAvatarURL string `json:"user_avatar_url"`
}

type BannedUser struct {
	ID            uint      `json:"id"`
	Username      string    `json:"username"`
	UserAvatarURL string    `json:"user_avatar_url"`
	BannedAt      time.Time `json:"banned_at"`
}

type BanUserRequest struct {
	Username string `json:"username"`
	Ban      bool   `json:"ban"`
}

// @Summary		Report an answer
// @Description	Report an answer given its ID
// @Tags			moderation
// @Param			id		path	string			true	"Answer id"
// @Param			report	body	ReportRequest	true	"Report cause"
// @Produce		json
// @Success		200	{object}	string
// @Failure		400	{object}	httputil.ApiError
// @Router			/moderation/report/{id} [post]
func ReportByIdHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	objectID := muxie.GetParam(w, "id")
	objID, err := strconv.ParseUint(objectID, 10, 0)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid proposal id")
		return
	}

	var req ReportRequest
	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid request body")
		slog.With("err", err).Error("failed to decode request body")
		return
	}

	if req.Cause == "" {
		httputil.WriteError(w, http.StatusBadRequest, "cause is required")
		return
	}

	user := middleware.MustGetUser(r)
	db := util.GetDb()

	err = util.SaveNewReport(db, uint(objID), req.Cause, user.ID)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to save report")
		slog.With("err", err).Error("failed to save report")
		return
	}

	httputil.WriteData(w, http.StatusOK, "Report saved successfully")
}

// @Summary		Get all reports
// @Description	Get all reports
// @Tags			moderation
// @Produce		json
// @Success		200	{object}	[]Report
// @Failure		400	{object}	httputil.ApiError
// @Router			/moderation/reports [get]
func GetReportsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if !middleware.GetAdmin(r) {
		httputil.WriteError(w, http.StatusForbidden, "you are not admin")
		return
	}

	db := util.GetDb()
	reports, err := util.GetAllReports(db)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to get reports")
		slog.With("err", err).Error("failed to get reports")
		return
	}

	returnResports := make([]Report, 0, len(reports))
	for _, report := range reports {

		user, err := util.GetUserByID(db, report.UserID)
		var username, avatarURL string
		if err != nil {
			slog.With("err", err).Error("failed to get user by id")
			username = "unknown"
			avatarURL = ""
		}

		username = user.Username
		avatarURL = util.GetPublicAvatarURL(user.ID)

		returnResports = append(returnResports, Report{
			ID:            report.ID,
			CreatedAt:     report.CreatedAt,
			UpdatedAt:     report.UpdatedAt,
			AnswerID:      report.AnswerID,
			Cause:         report.Cause,
			Username:      username,
			UserAvatarURL: avatarURL,
		})
	}

	httputil.WriteData(w, http.StatusOK, returnResports)
}

// @Summary		Get all banned users
// @Description	Get all banned users
// @Tags			moderation
// @Produce		json
// @Success		200	{object}	[]BannedUser
// @Failure		400	{object}	httputil.ApiError
// @Router			/moderation/ban [get]
func GetBannedHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if !middleware.GetAdmin(r) {
		httputil.WriteError(w, http.StatusForbidden, "you are not admin")
		return
	}

	db := util.GetDb()
	bannedUsers, err := util.GetBannedUsers(db)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to get banned users")
		slog.With("err", err).Error("failed to get banned users")
		return
	}

	returnBannedUsers := make([]BannedUser, 0, len(bannedUsers))
	for _, user := range bannedUsers {
		returnBannedUsers = append(returnBannedUsers, BannedUser{
			ID:            user.ID,
			Username:      user.Username,
			BannedAt:      user.UpdatedAt,
			UserAvatarURL: util.GetPublicAvatarURL(user.ID),
		})
	}

	httputil.WriteData(w, http.StatusOK, returnBannedUsers)
}

// @Summary		Ban or unban a user
// @Description	Ban or unban a user given its username
// @Tags			moderation
// @Param			banUser	body	BanUserRequest	true	"Ban or unban a user"
// @Produce		json
// @Success		200	{object}	string
// @Failure		400	{object}	httputil.ApiError
// @Router			/moderation/ban [post]
func BanUserHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if !middleware.GetAdmin(r) {
		httputil.WriteError(w, http.StatusForbidden, "you are not admin")
		return
	}

	var req BanUserRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid request body")
		slog.With("err", err).Error("failed to decode request body")
		return
	}

	if req.Username == "" {
		httputil.WriteError(w, http.StatusBadRequest, "username is required")
		return
	}

	db := util.GetDb()
	user, err := util.GetUserByUsername(db, req.Username)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to get user by username")
		slog.With("err", err).Error("failed to get user by username")
		return
	}
	if user == nil {
		httputil.WriteError(w, http.StatusBadRequest, "user not found")
		return
	}

	err = util.BanUnbanUser(db, req.Username, req.Ban)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to ban/unban user")
		slog.With("err", err).Error("failed to ban/unban user")
		return
	}

	if req.Ban {
		httputil.WriteData(w, http.StatusOK, "User banned successfully")
	} else {
		httputil.WriteData(w, http.StatusOK, "User unbanned successfully")
	}
}

// @Summary		Delete a report
// @Description	Delete a report given its ID
// @Tags			moderation
// @Param			id	path	string	true	"Answer id"
// @Produce		json
// @Success		204	{object}	nil
// @Failure		400	{object}	httputil.ApiError
// @Router			/moderation/report/{id} [delete]
func DeleteReportByIdHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if !middleware.GetAdmin(r) {
		httputil.WriteError(w, http.StatusForbidden, "you are not admin")
		return
	}

	objectID := muxie.GetParam(w, "id")
	objID, err := strconv.ParseUint(objectID, 10, 0)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid report id")
		return
	}

	db := util.GetDb()

	err = db.Delete(&models.Report{}, objID).Error
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to delete report")
		slog.With("err", err).Error("failed to delete report")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
