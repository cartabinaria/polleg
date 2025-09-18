package proposal

import (
	"net/http"

	"github.com/cartabinaria/auth/pkg/httputil"
	"github.com/cartabinaria/auth/pkg/middleware"
	"github.com/cartabinaria/polleg/models"
	"github.com/cartabinaria/polleg/util"
	"github.com/kataras/muxie"
	"golang.org/x/exp/slog"
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
	if !middleware.GetMember(req) {
		httputil.WriteError(res, http.StatusForbidden, "you are not a member")
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
	if !middleware.GetMember(req) {
		httputil.WriteError(res, http.StatusForbidden, "you are not a member")
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
