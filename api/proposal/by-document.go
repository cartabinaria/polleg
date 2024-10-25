package proposal

import (
	"fmt"
	"net/http"

	"github.com/cartabinaria/auth/pkg/httputil"
	"github.com/cartabinaria/auth/pkg/middleware"
	"github.com/cartabinaria/polleg/api"
	"github.com/cartabinaria/polleg/util"
	"github.com/kataras/muxie"
	"golang.org/x/exp/slog"
)

func ProposalByDocumentHandler(res http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodDelete:
		deleteProposalByDocumentHandler(res, req)
	case http.MethodGet:
		getProposalByDocumentHandler(res, req)
	default:
		httputil.WriteError(res, http.StatusMethodNotAllowed, "invalid method")
	}
}

func getProposalByDocumentHandler(res http.ResponseWriter, req *http.Request) {
	// Check method GET is used
	if req.Method != http.MethodGet {
		httputil.WriteError(res, http.StatusMethodNotAllowed, "invalid method")
		return
	}
	db := util.GetDb()
	docID := muxie.GetParam(res, "id")
	var questions []Proposal
	if err := db.Where(api.Question{Document: docID}).Find(&questions).Error; err != nil {
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

func deleteProposalByDocumentHandler(res http.ResponseWriter, req *http.Request) {
	if !middleware.GetAdmin(req) {
		httputil.WriteError(res, http.StatusForbidden, "you are not admin")
		return
	}
	if req.Method != http.MethodDelete {
		httputil.WriteError(res, http.StatusMethodNotAllowed, "invalid method")
		return
	}
	db := util.GetDb()
	docID := muxie.GetParam(res, "id")

	proposal := Proposal{Document: docID}

	if err := db.Where("document = ?", docID).Delete(&Proposal{}).Error; err != nil {
		fmt.Println()
		slog.Error("db query failed", "err", err)
		fmt.Println()
		httputil.WriteError(res, http.StatusInternalServerError, "db query failed")
		return
	}

	httputil.WriteData(res, http.StatusOK, proposal)
}
