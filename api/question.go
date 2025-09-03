package api

import (
	"net/http"
	"strconv"

	"github.com/cartabinaria/auth/pkg/httputil"
	"github.com/cartabinaria/auth/pkg/middleware"
	"github.com/cartabinaria/polleg/models"
	"github.com/cartabinaria/polleg/util"
	"github.com/kataras/muxie"
	"golang.org/x/exp/slog"
)

// @Summary		Get all answers given a question
// @Description	Given a question ID, return the question and all its answers
// @Tags			question
// @Param			id	path	string	true	"Answer id"
// @Produce		json
// @Success		200	{array}		models.QuestionResponse
// @Failure		400	{object}	httputil.ApiError
// @Router			/questions/{id} [get]
func GetQuestionHandler(res http.ResponseWriter, req *http.Request) {
	// Check method GET is used
	if req.Method != http.MethodGet {
		httputil.WriteError(res, http.StatusMethodNotAllowed, "invalid method")
		return
	}
	db := util.GetDb()
	rawQID := muxie.GetParam(res, "id")

	user, err := middleware.GetUser(req)
	requesterID := -1
	if err == nil {
		requesterID = int(user.ID)
	}
	isAdmin := middleware.GetAdmin(req)

	qID, err := strconv.ParseUint(rawQID, 10, 0)
	if err != nil {
		httputil.WriteError(res, http.StatusBadRequest, "invalid question id")
		return
	}

	var question models.Question
	if err := db.First(&question, uint(qID)).Error; err != nil {
		slog.Error("question not found", "err", err)
		httputil.WriteError(res, http.StatusNotFound, "question not found")
		return
	}
	var answers []models.Answer
	if err := db.Raw(ANSWERS_QUERY, question.ID).Scan(&answers).Error; err != nil {
		slog.Error("could not fetch answers", "err", err)
		httputil.WriteError(res, http.StatusInternalServerError, "could not fetch answers")
		return
	}
	answersIDs := []uint{}
	answersIndex := map[uint]int{}
	for i, answer := range answers {
		answersIDs = append(answersIDs, answer.ID)
		answersIndex[answer.ID] = i
	}
	var replies []models.Answer
	if err := db.Model(&models.Answer{}).Where("deleted_at is NULL AND parent IN ?", answersIDs).Find(&replies).Error; err != nil {
		slog.Error("could not fetch replies", "err", err)
		httputil.WriteError(res, http.StatusInternalServerError, "could not fetch replies")
		return
	}
	for _, reply := range replies {
		index := answersIndex[*reply.Parent]
		answers[index].Replies = append(answers[index].Replies, reply)
	}

	question.Answers = answers

	// recursively convert answers
	var responseAnswers []models.AnswerResponse
	for _, ans := range question.Answers {
		ans, err := ConvertAnswerToAPI(ans, isAdmin, requesterID)
		if err != nil {
			return
		}
		responseAnswers = append(responseAnswers, *ans)
	}

	httputil.WriteData(res, http.StatusOK,
		models.QuestionResponse{
			ID:        question.ID,
			CreatedAt: question.CreatedAt,
			UpdatedAt: question.UpdatedAt,
			Document:  question.Document,
			Start:     question.Start,
			End:       question.End,
			Answers:   responseAnswers,
		},
	)
}

// @Summary		Delete a question
// @Description	Given an andwer ID, delete the question
// @Tags			question
// @Param			id	path	string	true	"Question id"
// @Produce		json
// @Success		200	{object}	nil
// @Failure		400	{object}	httputil.ApiError
// @Router			/questions/{id} [delete]
func DelQuestionHandler(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodDelete {
		httputil.WriteError(res, http.StatusMethodNotAllowed, "invalid method")
		return
	}

	user := middleware.MustGetUser(req)
	if !user.Admin {
		httputil.WriteError(res, http.StatusForbidden, "only admins can delete questions")
		return
	}

	db := util.GetDb()
	rawAnsID := muxie.GetParam(res, "id")

	qID, err := strconv.ParseUint(rawAnsID, 10, 0)
	if err != nil {
		httputil.WriteError(res, http.StatusBadRequest, "invalid question id")
		return
	}

	if err := db.Delete(&models.Question{}, uint(qID)).Error; err != nil {
		slog.Error("something went wrong", "err", err)
		httputil.WriteError(res, http.StatusInternalServerError, "something went wrong")
		return
	}

	res.WriteHeader(http.StatusNoContent)
}
