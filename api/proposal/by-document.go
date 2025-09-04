package proposal

import (
	"fmt"
	"net/http"

	"github.com/cartabinaria/auth/pkg/httputil"
	"github.com/cartabinaria/auth/pkg/middleware"
	"github.com/cartabinaria/polleg/models"
	"github.com/cartabinaria/polleg/util"
	"github.com/kataras/muxie"
	"golang.org/x/exp/slog"
)

func ProposalByDocumentHandler(res http.ResponseWriter, req *http.Request) {
	if !middleware.GetAdmin(req) {
		httputil.WriteError(res, http.StatusForbidden, "you are not admin")
		return
	}

	switch req.Method {
	case http.MethodDelete:
		deleteProposalByDocumentHandler(res)
	case http.MethodGet:
		getProposalByDocumentHandler(res)
	default:
		httputil.WriteError(res, http.StatusMethodNotAllowed, "invalid method")
	}
}

func getProposalByDocumentHandler(res http.ResponseWriter) {
	db := util.GetDb()
	docID := muxie.GetParam(res, "id")

	var questions []Proposal
	if err := db.Where(models.Question{Document: docID}).Find(&questions).Error; err != nil {
		httputil.WriteError(res, http.StatusInternalServerError, "db query failed")
		return
	}
	if len(questions) == 0 {
		httputil.WriteError(res, http.StatusNotFound, "Document not found")
		return
	}

	httputil.WriteData(res, http.StatusOK, DocumentProposal{
		ID:        docID,
		Questions: questions,
	})
}

func deleteProposalByDocumentHandler(res http.ResponseWriter) {
	db := util.GetDb()
	docID := muxie.GetParam(res, "id")

	if err := db.Where("document = ?", docID).Delete(&Proposal{}).Error; err != nil {
		fmt.Println()
		slog.Error("db query failed", "err", err)
		fmt.Println()
		httputil.WriteError(res, http.StatusInternalServerError, "db query failed")
		return
	}

	res.WriteHeader(http.StatusNoContent)
}
