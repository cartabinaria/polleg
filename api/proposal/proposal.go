package proposal

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/cartabinaria/auth/pkg/httputil"
	"github.com/cartabinaria/auth/pkg/middleware"
	"github.com/cartabinaria/polleg/api"
	"github.com/cartabinaria/polleg/models"
	"github.com/cartabinaria/polleg/util"
	"github.com/kataras/muxie"
	"gorm.io/gorm/clause"
)

type DocumentProposal struct {
	ID        string     `json:"id"`
	Questions []Proposal `json:"questions"`
}

func ProposalHandler(res http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodPost:
		postProposal(res, req)
	case http.MethodGet:
		getAllProposalHandler(res, req)
	default:
		httputil.WriteError(res, http.StatusMethodNotAllowed, "invalid method")
	}
}

// Insert a question proposal (copy pasted from document.go)
func postProposal(res http.ResponseWriter, req *http.Request) {
	db := util.GetDb()

	// decode data
	var data api.PostDocumentRequest
	if err := json.NewDecoder(req.Body).Decode(&data); err != nil {
		httputil.WriteError(res, http.StatusBadRequest, "couldn't decode body")
		return
	}

	// save questions
	var questions []Proposal
	for _, coord := range data.Coords {
		q := Proposal{
			Document: data.ID,
			Start:    coord.Start,
			End:      coord.End,
		}
		questions = append(questions, q)
	}

	if err := db.Save(questions).Error; err != nil {
		httputil.WriteError(res, http.StatusInternalServerError, "couldn't create questions")
		return
	}

	httputil.WriteData(res, http.StatusOK, DocumentProposal{
		ID:        data.ID,
		Questions: questions,
	})
}

func groupByProperty[T any, K comparable](items []T, getProperty func(T) K) map[K][]T {
	grouped := make(map[K][]T)
	for _, item := range items {
		key := getProperty(item)
		grouped[key] = append(grouped[key], item)
	}
	return grouped
}

func getAllProposalHandler(res http.ResponseWriter, req *http.Request) {
	if !middleware.GetAdmin(req) {
		httputil.WriteError(res, http.StatusForbidden, "you are not admin")
		return
	}

	db := util.GetDb()
	var questions []Proposal
	if err := db.Where(Proposal{}).Find(&questions).Error; err != nil {
		httputil.WriteError(res, http.StatusInternalServerError, "db query failed")
		return
	}
	if len(questions) == 0 {
		httputil.WriteError(res, http.StatusInternalServerError, "No proposal found")
		return
	}

	// group proposal by the document
	groupedByDoc := groupByProperty(questions, func(p Proposal) string {
		return p.Document
	})

	docProps := []DocumentProposal{}
	for doc, group := range groupedByDoc {
		var qs []Proposal
		for _, proposal := range group {
			q := Proposal{
				Document: doc,
				Start:    proposal.Start,
				End:      proposal.End,
			}
			qs = append(qs, q)
		}
		data := DocumentProposal{
			ID:        doc,
			Questions: qs,
		}
		docProps = append(docProps, data)
	}

	if len(docProps) == 0 {
		httputil.WriteError(res, http.StatusNotFound, "Proposal not found")
		return
	}

	httputil.WriteData(res, http.StatusOK, docProps)
}

// @Summary		Approve a proposal
// @Description	Approve a proposal given its id
// @Tags			answer
// @Param			answerReq	body	nil
// @Produce		json
// @Success		200	{object}	models.AnswerResponse
// @Failure		400	{object}	httputil.ApiError
// @Router			/proposal/:id/approve [post]
func ApproveProposalHandler(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		httputil.WriteError(res, http.StatusMethodNotAllowed, "invalid method")
		return
	}

	if !middleware.GetAdmin(req) {
		httputil.WriteError(res, http.StatusForbidden, "you are not admin")
		return
	}

	rawID := muxie.GetParam(res, "id")
	proposalID, err := strconv.Atoi(rawID)
	if err != nil {
		httputil.WriteError(res, http.StatusBadRequest, "invalid id")
		return
	}

	db := util.GetDb()
	tx := db.Begin()
	if err := tx.Error; err != nil {
		httputil.WriteError(res, http.StatusInternalServerError, "couldn't start transaction")
		return
	}

	var proposal Proposal
	if err := tx.Clauses(clause.Returning{}).Delete(&proposal, proposalID).Error; err != nil {
		tx.Rollback()
		httputil.WriteError(res, http.StatusNotFound, "proposal not found")
		return
	}

	question := models.Question{
		Document: proposal.Document,
		Start:    proposal.Start,
		End:      proposal.End,
	}

	if err := db.Create(&question).Error; err != nil {
		tx.Rollback()
		httputil.WriteError(res, http.StatusInternalServerError, "couldn't create question")
		return
	}

	if err := tx.Commit().Error; err != nil {
		httputil.WriteError(res, http.StatusInternalServerError, "couldn't commit transaction")
		return
	}

	httputil.WriteData(res, http.StatusOK, proposal)
}
