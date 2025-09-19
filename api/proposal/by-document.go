package proposal

import (
	"net/http"

	"github.com/cartabinaria/auth/pkg/httputil"
	"github.com/cartabinaria/auth/pkg/middleware"
	"github.com/cartabinaria/polleg/models"
	"github.com/cartabinaria/polleg/util"
	"github.com/kataras/muxie"
	"golang.org/x/exp/slog"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// @Summary		Get proposals by document id
// @Description	Get all proposals for a document, given its ID
// @Tags			proposal
// @Param			id	path	string	true	"Document id"
// @Produce		json
// @Success		200	{object}	Proposal
// @Failure		400	{object}	httputil.ApiError
// @Router			/proposals/document/{id} [get]
func GetProposalByDocumentHandler(res http.ResponseWriter, req *http.Request) {
	if !middleware.GetMember(req) && !middleware.GetAdmin(req) {
		httputil.WriteError(res, http.StatusForbidden, "you are not a member or admin")
		return
	}

	db := util.GetDb()
	docID := muxie.GetParam(res, "id")

	var questions []models.Proposal
	if err := db.Where("document = ?", docID).Find(&questions).Error; err != nil {
		slog.Error("db query failed", "err", err)
		httputil.WriteError(res, http.StatusInternalServerError, "db query failed")
		return
	}

	httputil.WriteData(res, http.StatusOK, DocumentProposal{
		ID:           docID,
		DocumentPath: "", // In this endpoint we don't return the document path
		Questions:    dbProposalsToProposals(db, questions),
	})
}

// @Summary		Delete all proposals for a document
// @Description	Given a document ID, delete all its proposals
// @Tags			proposal
// @Param			id	path	string	true	"Document id"
// @Produce		json
// @Success		200	{object}	nil
// @Failure		400	{object}	httputil.ApiError
// @Router			/proposals/document/{id} [delete]
func DeleteProposalByDocumentHandler(res http.ResponseWriter, req *http.Request) {
	if !middleware.GetMember(req) && !middleware.GetAdmin(req) {
		httputil.WriteError(res, http.StatusForbidden, "you are not a member or admin")
		return
	}

	db := util.GetDb()
	docID := muxie.GetParam(res, "id")

	if err := db.Where("document = ?", docID).Delete(&models.Proposal{}).Error; err != nil {
		slog.With("err", err).Error("db query failed")
		httputil.WriteError(res, http.StatusInternalServerError, "db query failed")
		return
	}

	res.WriteHeader(http.StatusNoContent)
}

// @Summary		Approve all proposals for a document
// @Description	Given a document ID, approve all its proposals
// @Tags			proposal
// @Param			id	path	string	true	"Document id"
// @Produce		json
// @Success		200	{object}	nil
// @Failure		400	{object}	httputil.ApiError
// @Router			/proposals/document/{id}/approve [delete]
func ApproveProposalByDocumentHandler(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		httputil.WriteError(res, http.StatusMethodNotAllowed, "invalid method")
		return
	}

	if !middleware.GetMember(req) && !middleware.GetAdmin(req) {
		httputil.WriteError(res, http.StatusForbidden, "you are not a member or admin")
		return
	}

	docID := muxie.GetParam(res, "id")
	db := util.GetDb()

	var proposals []models.Proposal
	var questions []models.Question

	err := db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Clauses(clause.Returning{}).Where("document_id = ?", docID).Delete(&proposals).Error; err != nil {
			slog.Error("error while deleting proposals", "proposals", proposals, "err", err)
			return err
		}

		for _, proposal := range proposals {
			questions = append(questions, models.Question{
				Document: proposal.DocumentID,
				Start:    proposal.Start,
				End:      proposal.End,
				UserID:   proposal.UserID,
			})
		}

		if err := tx.Create(&questions).Error; err != nil {
			slog.Error("error while creating the questions", "questions", questions, "err", err)
			return err
		}

		return nil
	})
	if err != nil {
		slog.With("err", err).Error("transaction failed")
		httputil.WriteError(res, http.StatusInternalServerError, "transaction failed")
		return
	}
}
