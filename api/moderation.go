package api

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/cartabinaria/auth/pkg/httputil"
	"github.com/cartabinaria/auth/pkg/middleware"
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
	Username      string `json:"user_id"`
	UserAvatarURL string `json:"user_avatar_url"`
}

// @Summary		Report an answer
// @Description	Report an answer given its ID
// @Tags			moderation
// @Param			id	path	string	true	"Answer id"
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

	returnResports := make([]Report, len(reports))
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
