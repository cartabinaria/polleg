package api

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/cartabinaria/auth/pkg/httputil"
	"github.com/cartabinaria/auth/pkg/middleware"
	"github.com/cartabinaria/polleg/models"
	"github.com/cartabinaria/polleg/util"
	"github.com/kataras/muxie"
)

type Document struct {
	ID        string     `json:"id"`
	Questions []Question `json:"questions"`
}

type Coord struct {
	Start uint32 `json:"start"`
	End   uint32 `json:"end"`
}

type PostDocumentRequest struct {
	ID     string  `json:"id"`
	Coords []Coord `json:"coords"`
}

// @Summary		Insert a new document
// @Description	Insert a new document with all the questions initialised
// @Tags			document
// @Param			docRequest	body	PostDocumentRequest	true	"Doc request body"
// @Produce		json
// @Success		200	{object}	Document
// @Failure		400	{object}	httputil.ApiError
// @Router			/documents [post]
func PostDocumentHandler(res http.ResponseWriter, req *http.Request) {
	// Check method POST is used
	if req.Method != http.MethodPost {
		httputil.WriteError(res, http.StatusMethodNotAllowed, "invalid method")
		return
	}

	// Only members of the staff can add a document
	if !middleware.GetMember(req) && !middleware.GetAdmin(req) {
		httputil.WriteError(res, http.StatusForbidden, "you are not a member or admin")
		return
	}
	db := util.GetDb()
	user := middleware.MustGetUser(req)
	_, err := util.GetOrCreateUserByID(db, user.ID, user.Username)
	if err != nil {
		slog.With("user", user, "err", err).Error("error while getting or creating the user-alias association")
		httputil.WriteError(res, http.StatusBadRequest, "could not insert the answer")
		return
	}

	// decode data
	var data PostDocumentRequest
	if err := json.NewDecoder(req.Body).Decode(&data); err != nil {
		httputil.WriteError(res, http.StatusBadRequest, "couldn't decode body")
		return
	}

	// save questions
	var questions []models.Question
	for _, coord := range data.Coords {
		q := models.Question{
			Document: data.ID,
			Start:    coord.Start,
			End:      coord.End,
			UserID:   uint(user.ID),
		}
		questions = append(questions, q)
	}

	if err := db.Save(questions).Error; err != nil {
		httputil.WriteError(res, http.StatusInternalServerError, "couldn't create questions")
		return
	}

	httputil.WriteData(res, http.StatusOK, Document{
		ID:        data.ID,
		Questions: dbQuestionsToQuestions(questions),
	})
}

// @Summary		Get a document's divisions
// @Description	Given a document's ID, return all the questions
// @Tags			document
// @Param			id	path	string	true	"document id"
// @Produce		json
// @Success		200	{object}	Document
// @Failure		400	{object}	httputil.ApiError
// @Router			/documents/{id} [get]
func GetDocumentHandler(res http.ResponseWriter, req *http.Request) {
	// Check method GET is used
	if req.Method != http.MethodGet {
		httputil.WriteError(res, http.StatusMethodNotAllowed, "invalid method")
		return
	}
	db := util.GetDb()
	docID := muxie.GetParam(res, "id")
	var dbQuestions []models.Question
	if err := db.Where(models.Question{Document: docID}).Find(&dbQuestions).Error; err != nil {
		httputil.WriteError(res, http.StatusInternalServerError, "db query failed")
		return
	}
	if len(dbQuestions) == 0 {
		httputil.WriteError(res, http.StatusNotFound, "Document not found")
		return
	}

	httputil.WriteData(res, http.StatusOK, Document{
		ID:        docID,
		Questions: dbQuestionsToQuestions(dbQuestions),
	})
}

func dbQuestionsToQuestions(q []models.Question) []Question {
	questions := make([]Question, len(q))
	for i, question := range q {
		questions[i] = Question{
			ID:        question.ID,
			CreatedAt: question.CreatedAt,
			UpdatedAt: question.UpdatedAt,

			Document: question.Document,
			Start:    question.Start,
			End:      question.End,
			Answers:  nil,
		}
	}
	return questions
}
