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
	"golang.org/x/exp/slog"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type DocumentProposal struct {
	ID           string     `json:"id"`
	DocumentPath string     `json:"document_path,omitempty"`
	Questions    []Proposal `json:"questions"`
}

type PostDocumentProposalRequest struct {
	// ID is calculated with sha256sum of the document path
	ID string `json:"id"`
	// In order to list all available documents with proposals, we have to
	// store the document path too. We can't use the ID because it's irreversible.
	DocumentPath string      `json:"document_path,omitempty"`
	Coords       []api.Coord `json:"coords"`
}

// @Summary		Insert a new proposal
// @Description	Insert a new proposal for a document
// @Tags			proposal
// @Param			proposalReq	body	DocumentProposal	true	"Proposal data to insert"
// @Produce		json
// @Success		200	{object}	DocumentProposal
// @Failure		400	{object}	httputil.ApiError
// @Router			/proposals [post]
func PostProposalHandler(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		httputil.WriteError(res, http.StatusMethodNotAllowed, "invalid method")
		return
	}

	// decode data
	var data PostDocumentProposalRequest
	if err := json.NewDecoder(req.Body).Decode(&data); err != nil {
		httputil.WriteError(res, http.StatusBadRequest, "couldn't decode body")
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

	// save questions
	var questions []models.Proposal
	for _, coord := range data.Coords {
		q := models.Proposal{
			DocumentID:   data.ID,
			DocumentPath: data.DocumentPath,
			Start:        coord.Start,
			End:          coord.End,
			UserID:       uint(user.ID),
		}
		questions = append(questions, q)
	}

	if err := db.Save(questions).Error; err != nil {
		httputil.WriteError(res, http.StatusInternalServerError, "couldn't create questions")
		return
	}

	httputil.WriteData(res, http.StatusOK, dbProposalsToProposals(db, questions))
}

func groupByProperty[T any, K comparable](items []T, getProperty func(T) K) map[K][]T {
	grouped := make(map[K][]T)
	for _, item := range items {
		key := getProperty(item)
		grouped[key] = append(grouped[key], item)
	}
	return grouped
}

// @Summary		Get all proposals
// @Description	Get all proposals
// @Tags			proposal
// @Produce		json
// @Success		200	{object}	[]DocumentProposal
// @Failure		400	{object}	httputil.ApiError
// @Router			/proposals [get]
func GetAllProposalsHandler(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		httputil.WriteError(res, http.StatusMethodNotAllowed, "invalid method")
		return
	}

	if !middleware.GetMember(req) && !middleware.GetAdmin(req) {
		httputil.WriteError(res, http.StatusForbidden, "you are not a member or admin")
		return
	}

	db := util.GetDb()
	var dbProposals []models.Proposal
	if err := db.Find(&dbProposals).Error; err != nil {
		httputil.WriteError(res, http.StatusInternalServerError, "db query failed")
		return
	}

	proposals := dbProposalsToProposals(db, dbProposals)

	// group proposal by the document
	groupedByDoc := groupByProperty(proposals, func(p Proposal) string {
		return p.DocumentID
	})

	var docProps []DocumentProposal
	for doc, group := range groupedByDoc {
		data := DocumentProposal{
			ID:           doc,
			DocumentPath: group[0].DocumentPath,
			Questions:    group,
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
// @Produce		json
// @Success		200	{object}	models.Question
// @Failure		400	{object}	httputil.ApiError
// @Router			/proposals/{id}/approve [post]
func ApproveProposalHandler(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		httputil.WriteError(res, http.StatusMethodNotAllowed, "invalid method")
		return
	}

	if !middleware.GetMember(req) && !middleware.GetAdmin(req) {
		httputil.WriteError(res, http.StatusForbidden, "you are not a member or admin")
		return
	}

	rawID := muxie.GetParam(res, "id")
	proposalID, err := strconv.Atoi(rawID)
	if err != nil {
		httputil.WriteError(res, http.StatusBadRequest, "invalid id")
		return
	}

	db := util.GetDb()
	user := middleware.MustGetUser(req)
	_, err = util.GetOrCreateUserByID(db, user.ID, user.Username)

	var proposal models.Proposal
	var question models.Question

	err = db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Clauses(clause.Returning{}).Delete(&proposal, proposalID).Error; err != nil {
			slog.Error("error while deleting proposal", "proposal", proposal, "err", err)
			return err
		}

		// Create question
		question = models.Question{
			Document: proposal.DocumentID,
			Start:    proposal.Start,
			End:      proposal.End,
			UserID:   uint(user.ID),
		}

		if err := tx.Create(&question).Error; err != nil {
			slog.Error("error while creating the question", "question", question, "err", err)
			return err
		}

		return nil
	})

	if err != nil {
		httputil.WriteError(res, http.StatusBadRequest, "could not approve proposal")
		return
	}

	httputil.WriteData(res, http.StatusOK, api.Question{
		ID:        question.ID,
		CreatedAt: question.CreatedAt,
		UpdatedAt: question.UpdatedAt,
		Document:  question.Document,
		Start:     question.Start,
		End:       question.End,
		Answers:   []api.Answer{},
	})
}
