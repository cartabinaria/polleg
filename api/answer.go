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
	"golang.org/x/exp/slog"
	"gorm.io/gorm"
)

var (
	VOTES_QUERY = fmt.Sprintf(`
  SELECT votes.answer,
				 COUNT(CASE votes.vote WHEN %d THEN 1 ELSE NULL END) as upvotes,
				 COUNT(CASE votes.vote WHEN %d THEN 1 ELSE NULL END) as downvotes
  FROM votes
  GROUP BY Answer
`, VoteUp, VoteDown)
	ANSWERS_QUERY = fmt.Sprintf(`
  SELECT *
  FROM answers
  LEFT JOIN (%s) vote_counts ON vote_counts.answer = answers.id
  WHERE answers.deleted_at is NULL
		AND answers.parent is NULL
		AND answers.question = ?
`, VOTES_QUERY)
)

func ConvertAnswerToAPI(answer models.Answer, isAdmin bool, requesterID int) (*models.AnswerResponse, error) {
	db := util.GetDb()
	usr, err := util.GetUserByID(db, answer.UserId)
	if err != nil {
		return nil, err
	}

	var avatar, username, content string

	if answer.State != models.AnswerStateVisible {
		username = "[deleted]"
		avatar = util.DeletedURL
		content = "[deleted]"
	} else if answer.Anonymous {
		avatar = util.GenerateAnonymousAvatar(usr.Alias)
		username = usr.Alias
		content = answer.Content
	} else {
		avatar = fmt.Sprintf("https://avatars.githubusercontent.com/u/%d?v=4", usr.ID)
		username = usr.Username
		content = answer.Content
	}

	var voteValue models.VoteValue
	var vote models.Vote
	err = db.Where("answer = ? AND user_id = ?", answer.ID, requesterID).First(&vote).Error
	if err != nil {
		if err != gorm.ErrRecordNotFound {
			return nil, err
		} else {
			voteValue = VoteNone
		}
	} else {
		voteValue = models.VoteValue(vote.Vote)
	}

	// recursively convert replies
	var replies []models.AnswerResponse
	for _, reply := range answer.Replies {
		reply, err := ConvertAnswerToAPI(reply, isAdmin, requesterID)
		if err != nil {
			return nil, err
		}
		replies = append(replies, *reply)
	}

	return &models.AnswerResponse{
		ID:            answer.ID,
		CreatedAt:     answer.CreatedAt,
		UpdatedAt:     answer.UpdatedAt,
		Question:      answer.Question,
		Parent:        answer.Parent,
		User:          username,
		UserAvatarURL: avatar,
		Content:       content,
		Upvotes:       answer.Upvotes,
		Downvotes:     answer.Downvotes,
		Replies:       replies,
		CanIDelete:    isAdmin || int(answer.UserId) == requesterID,
		IVoted:        voteValue,
		EditedByAdmin: answer.EditedByAdmin,
	}, nil

}

// @Summary		Insert a new answer
// @Description	Insert a new answer under a question
// @Tags			answer
// @Param			answerReq	body	models.PostAnswerRequest	true	"Answer data to insert"
// @Produce		json
// @Success		200	{object}	models.AnswerResponse
// @Failure		400	{object}	httputil.ApiError
// @Router			/answers [post]
func PostAnswerHandler(res http.ResponseWriter, req *http.Request) {
	// Check method POST is used
	if req.Method != http.MethodPost {
		httputil.WriteError(res, http.StatusMethodNotAllowed, "invalid method")
		return
	}
	db := util.GetDb()
	user := middleware.MustGetUser(req)

	var ans models.PostAnswerRequest
	err := json.NewDecoder(req.Body).Decode(&ans)
	if err != nil {
		httputil.WriteError(res, http.StatusBadRequest, fmt.Sprintf("decode error: %v", err))
		return
	}

	var quest models.Question
	if err := db.First(&quest, ans.Question).Error; err != nil {
		httputil.WriteError(res, http.StatusBadRequest, "the referenced question does not exist")
		return
	}

	if ans.Parent != nil {
		var Parent models.Answer
		if err = db.First(&Parent, ans.Parent).Error; err != nil {
			httputil.WriteError(res, http.StatusBadRequest, "the referenced parent does not exist")
			return
		}
		if Parent.Question != quest.ID {
			httputil.WriteError(res, http.StatusBadRequest, "mismatch between parent question and this question")
			return
		}
	}

	// TODO: upvotes and downvotes should really be just the result of a
	// COUNT() aggregator on the votes table
	answer := models.Answer{
		Question:  ans.Question,
		Parent:    ans.Parent,
		UserId:    user.ID,
		Content:   ans.Content,
		Upvotes:   0,
		Downvotes: 0,
		Anonymous: ans.Anonymous,
	}

	err = db.Create(&answer).Error
	if err != nil {
		slog.Error("error while creating the answer", "answer", answer, "err", err)
		httputil.WriteError(res, http.StatusBadRequest, "could not insert the answer")
		return
	}

	usr, err := util.GetOrCreateUserByID(db, user.ID, user.Username)
	if err != nil {
		slog.Error("error while getting or creating the user-alias association", "user", user, "err", err)
		httputil.WriteError(res, http.StatusBadRequest, "could not insert the answer")
		return
	}

	var avatar, username string

	if ans.Anonymous {
		avatar = util.GenerateAnonymousAvatar(usr.Alias)
		username = usr.Alias
	} else {
		avatar = user.AvatarUrl
		username = user.Username
	}

	httputil.WriteData(res, http.StatusOK,
		models.AnswerResponse{
			ID:            answer.ID,
			CreatedAt:     answer.CreatedAt,
			UpdatedAt:     answer.UpdatedAt,
			Question:      answer.Question,
			Parent:        answer.Parent,
			User:          username,
			UserAvatarURL: avatar,
			Content:       answer.Content,
			Upvotes:       answer.Upvotes,
			Downvotes:     answer.Downvotes,
			CanIDelete:    true,
			IVoted:        0,
		})
}

// @Summary		Delete an answer
// @Description	Given an andwer ID, delete the answer
// @Tags			answer
// @Param			id	path	string	true	"Answer id"
// @Produce		json
// @Success		200	{object}	nil
// @Failure		400	{object}	httputil.ApiError
// @Router			/answers/{id} [delete]
func DelAnswerHandler(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodDelete {
		httputil.WriteError(res, http.StatusMethodNotAllowed, "invalid method")
		return
	}

	user := middleware.MustGetUser(req)
	db := util.GetDb()
	rawAnsID := muxie.GetParam(res, "id")

	aID, err := strconv.ParseUint(rawAnsID, 10, 0)
	if err != nil {
		httputil.WriteError(res, http.StatusBadRequest, "invalid question id")
		return
	}

	var answer models.Answer
	if err := db.First(&answer, uint(aID)).Error; err != nil {
		slog.Error("answer not found", "err", err)
		httputil.WriteError(res, http.StatusNotFound, "answer not found")
		return
	}

	if !user.Admin && answer.UserId != user.ID {
		slog.Error("you are not an admin or the owner of the answer", "err", err)
		httputil.WriteError(res, http.StatusUnauthorized, "you are not an admin or the owner of the answer")
		return
	}

	if answer.State == models.AnswerStateDeletedByUser || answer.State == models.AnswerStateDeletedByAdmin {
		httputil.WriteError(res, http.StatusBadRequest, "the answer has already been deleted")
		return
	}

	if user.ID != answer.UserId && user.Admin {
		answer.State = models.AnswerStateDeletedByAdmin
	} else {
		answer.State = models.AnswerStateDeletedByUser
	}

	if err := db.Save(&answer).Error; err != nil {
		slog.Error("couldn't delete answer", "err", err)
		httputil.WriteError(res, http.StatusInternalServerError, "couldn't delete answer")
		return
	}

	res.WriteHeader(http.StatusNoContent)
}

// @Summary		Update an answer
// @Description	Given an andwer ID, update the answer
// @Tags			answer
// @Param			id	path	string	true	"Answer id"
// @Produce		json
// @Success		200	{object}	nil
// @Failure		400	{object}	httputil.ApiError
// @Router			/answers/{id} [patch]
func UpdateAnswerHandler(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPatch {
		httputil.WriteError(res, http.StatusMethodNotAllowed, "invalid method")
		return
	}

	user := middleware.MustGetUser(req)
	db := util.GetDb()
	rawAnsID := muxie.GetParam(res, "id")

	aID, err := strconv.ParseUint(rawAnsID, 10, 0)
	if err != nil {
		httputil.WriteError(res, http.StatusBadRequest, "invalid question id")
		return
	}

	var body models.UpdateAnswerRequest
	err = json.NewDecoder(req.Body).Decode(&body)
	if err != nil {
		httputil.WriteError(res, http.StatusBadRequest, fmt.Sprintf("decode error: %v", err))
		return
	}

	var answer models.Answer
	if err := db.First(&answer, uint(aID)).Error; err != nil {
		slog.Error("answer not found", "err", err)
		httputil.WriteError(res, http.StatusNotFound, "answer not found")
		return
	}

	if !user.Admin && answer.UserId != user.ID {
		slog.Error("you are not an admin or the owner of the answer", "err", err)
		httputil.WriteError(res, http.StatusUnauthorized, "you are not an admin or the owner of the answer")
		return
	}

	if answer.State == models.AnswerStateDeletedByUser || answer.State == models.AnswerStateDeletedByAdmin {
		httputil.WriteError(res, http.StatusBadRequest, "you cannot update a deleted answer")
		return
	}

	answer.Content = body.Content

	if answer.UserId != user.ID && user.Admin {
		answer.EditedByAdmin = true
	}

	if err := db.Save(&answer).Error; err != nil {
		slog.Error("couldn't update answer", "err", err)
		httputil.WriteError(res, http.StatusInternalServerError, "couldn't update answer")
		return
	}

	responseData, err := ConvertAnswerToAPI(answer, user.Admin, int(user.ID))
	if err != nil {
		slog.Error("couldn't update answer", "err", err)
		httputil.WriteError(res, http.StatusInternalServerError, "couldn't update answer")
		return
	}

	httputil.WriteData(res, http.StatusOK, responseData)
}
