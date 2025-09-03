package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/cartabinaria/auth/pkg/httputil"
	"github.com/cartabinaria/auth/pkg/middleware"
	"github.com/cartabinaria/polleg/models"
	"github.com/cartabinaria/polleg/util"
	"github.com/kataras/muxie"
	"gorm.io/gorm/clause"
)

const (
	VoteUp   models.VoteValue = 1
	VoteNone models.VoteValue = 0
	VoteDown models.VoteValue = -1
)

// get given vote to an answer
func GetUserVote(res http.ResponseWriter, req *http.Request) {
	// Check method GET is used
	if req.Method != http.MethodGet {
		httputil.WriteError(res, http.StatusMethodNotAllowed, "invalid method")
		return
	}
	db := util.GetDb()
	user := middleware.MustGetUser(req)

	rawAnsID := muxie.GetParam(res, "id")
	ansID, err := strconv.ParseUint(rawAnsID, 10, 0)
	if err != nil {
		httputil.WriteError(res, http.StatusBadRequest, "invalid answer id")
		return
	}

	var vote models.Vote
	if err = db.First(&vote, "answer = ? and \"user\" = ?", ansID, user.ID).Error; err != nil {
		httputil.WriteError(res, http.StatusBadRequest, "the referenced vote does not exist")
		return
	}

	httputil.WriteData(res, http.StatusOK, models.VoteResponse{
		Answer:    vote.Answer,
		User:      user.Username,
		Vote:      int8(vote.Vote),
		CreatedAt: vote.CreatedAt,
		UpdatedAt: vote.UpdatedAt,
	})
}

// @Summary		Insert a vote
// @Description	Insert a new vote on a answer
// @Tags			vote
// @Produce		json
// @Param			id	path		string	true	"code query parameter"
// @Success		200	{object}	models.VoteResponse
// @Failure		400	{object}	httputil.ApiError
// @Router			/answer/{id}/vote [post]
func PostVote(res http.ResponseWriter, req *http.Request) {
	if req.Method == http.MethodGet {
		GetUserVote(res, req)
		return
	}
	// Check method POST is used
	if req.Method != http.MethodPost {
		httputil.WriteError(res, http.StatusMethodNotAllowed, "invalid method")
		return
	}
	db := util.GetDb()
	user := middleware.MustGetUser(req)

	rawAnsID := muxie.GetParam(res, "id")
	ansID, err := strconv.ParseUint(rawAnsID, 10, 0)
	if err != nil {
		httputil.WriteError(res, http.StatusBadRequest, "invalid answer id")
		return
	}

	var v models.PostVoteRequest
	err = json.NewDecoder(req.Body).Decode(&v)
	if err != nil {
		httputil.WriteError(res, http.StatusBadRequest, fmt.Sprintf("decode error: %v", err))
		return
	}

	var ans models.Answer
	if err = db.First(&ans, ansID).Error; err != nil {
		httputil.WriteError(res, http.StatusBadRequest, "the referenced answer does not exist")
		return
	}
	if ans.Parent != nil {
		httputil.WriteError(res, http.StatusBadRequest, "cannot vote a reply to an answer")
		return
	}

	vote := models.Vote{
		Answer: ans.ID,
		UserId: user.ID,
		Vote:   int8(v.Vote),
	}
	switch v.Vote {
	case VoteUp, VoteDown:
		// If a vote already exists, and
		err := db.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "answer"}, {Name: "user_id"}},
			DoUpdates: clause.AssignmentColumns([]string{"vote"}),
		}).Create(&vote).Error
		if err != nil {
			httputil.WriteError(res, http.StatusInternalServerError, "could not update your vote")
			return
		}

	case VoteNone:
		result := db.Where("answer = ? AND user_id = ?", ans.ID, user.ID).Delete(&models.Vote{})
		if result.Error != nil {
			httputil.WriteError(res, http.StatusInternalServerError, "could not delete the previous vote")
			return
		}
		if result.RowsAffected == 0 {
			httputil.WriteError(res, http.StatusNotFound, "no vote found to delete")
			return
		}

	default:
		httputil.WriteError(res, http.StatusBadRequest, "the vote value must be either 1, -1 or 0")
		return
	}

	httputil.WriteData(res, http.StatusOK, models.VoteResponse{
		Answer:    vote.Answer,
		User:      user.Username,
		Vote:      int8(vote.Vote),
		CreatedAt: vote.CreatedAt,
		UpdatedAt: vote.UpdatedAt,
	})
}
